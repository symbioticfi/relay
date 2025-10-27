package api_server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-errors/errors"
)

// mockFlusher implements http.Flusher for testing
type mockFlusher struct {
	flushCount int
}

func (m *mockFlusher) Flush() {
	m.flushCount++
}

// mockResponseWriter implements http.ResponseWriter and http.Flusher for testing
type mockResponseWriter struct {
	mockFlusher

	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestSSEResponseWriter_CompleteMessages(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Write complete JSON messages
	input := `{"id":1,"name":"test1"}
{"id":2,"name":"test2"}
`
	n, err := writer.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(input) {
		t.Errorf("Expected to write %d bytes, got %d", len(input), n)
	}

	// Check output is in SSE format
	output := mock.body.String()
	expected := "data: {\"id\":1,\"name\":\"test1\"}\n\ndata: {\"id\":2,\"name\":\"test2\"}\n\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}

	// Check that flush was called for each message
	if mock.flushCount != 2 {
		t.Errorf("Expected 2 flushes, got %d", mock.flushCount)
	}

	// Verify headers were set correctly
	if mock.header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type: text/event-stream, got: %s", mock.header.Get("Content-Type"))
	}
}

func TestSSEResponseWriter_PartialWrites(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Simulate json.Encoder behavior: writes in chunks without newline
	chunk1 := `{"id":1,"name":"very_long`
	chunk2 := `_field_that_spans_multiple_writes","data":"more`
	chunk3 := `_data_here"}` + "\n"

	// Write chunks one by one
	n1, err := writer.Write([]byte(chunk1))
	if err != nil {
		t.Fatalf("Write chunk1 failed: %v", err)
	}
	if n1 != len(chunk1) {
		t.Errorf("Expected to write %d bytes, got %d", len(chunk1), n1)
	}

	// After first write, no output should be produced (no complete line yet)
	if mock.body.Len() > 0 {
		t.Errorf("Expected no output after partial write, got: %s", mock.body.String())
	}

	n2, err := writer.Write([]byte(chunk2))
	if err != nil {
		t.Fatalf("Write chunk2 failed: %v", err)
	}
	if n2 != len(chunk2) {
		t.Errorf("Expected to write %d bytes, got %d", len(chunk2), n2)
	}

	// Still no output (no newline yet)
	if mock.body.Len() > 0 {
		t.Errorf("Expected no output after second partial write, got: %s", mock.body.String())
	}

	n3, err := writer.Write([]byte(chunk3))
	if err != nil {
		t.Fatalf("Write chunk3 failed: %v", err)
	}
	if n3 != len(chunk3) {
		t.Errorf("Expected to write %d bytes, got %d", len(chunk3), n3)
	}

	// Now we should have complete output
	output := mock.body.String()
	expectedJSON := `{"id":1,"name":"very_long_field_that_spans_multiple_writes","data":"more_data_here"}`
	expected := fmt.Sprintf("data: %s\n\n", expectedJSON)
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}

	// Should have flushed once (after complete line)
	if mock.flushCount != 1 {
		t.Errorf("Expected 1 flush, got %d", mock.flushCount)
	}
}

func TestSSEResponseWriter_EmptyLines(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Write with empty lines
	input := `{"id":1}

{"id":2}
`
	n, err := writer.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(input) {
		t.Errorf("Expected to write %d bytes, got %d", len(input), n)
	}

	// Empty lines should be skipped
	output := mock.body.String()
	expected := "data: {\"id\":1}\n\ndata: {\"id\":2}\n\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

func TestSSEResponseWriter_MultipleMessagesWithPartials(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Write complete message followed by partial
	input1 := `{"id":1}
{"id":2,"partial`
	n1, err := writer.Write([]byte(input1))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n1 != len(input1) {
		t.Errorf("Expected to write %d bytes, got %d", len(input1), n1)
	}

	// Check that first message was output
	output1 := mock.body.String()
	if !strings.Contains(output1, `data: {"id":1}`) {
		t.Errorf("Expected first message in output, got: %s", output1)
	}
	if strings.Contains(output1, `partial`) {
		t.Errorf("Partial message should not be in output yet, got: %s", output1)
	}

	// Complete the partial message
	input2 := `":true}
`
	n2, err := writer.Write([]byte(input2))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n2 != len(input2) {
		t.Errorf("Expected to write %d bytes, got %d", len(input2), n2)
	}

	// Now both messages should be present
	output2 := mock.body.String()
	if !strings.Contains(output2, `data: {"id":2,"partial":true}`) {
		t.Errorf("Expected second message in output, got: %s", output2)
	}
}

func TestSSEResponseWriter_BufferOverflow(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Create a message larger than 5MB (the maxBufferSize) without newline
	largeInput := make([]byte, 5*1024*1024+1)
	for i := range largeInput {
		largeInput[i] = 'x'
	}

	_, err := writer.Write(largeInput)
	if err == nil {
		t.Fatal("Expected buffer overflow error, got nil")
	}
	if !strings.Contains(err.Error(), "buffer overflow") {
		t.Errorf("Expected buffer overflow error, got: %v", err)
	}
}

func TestSSEResponseWriter_WriteError(t *testing.T) {
	// Create a mock that fails on write
	mock := &failingResponseWriter{
		header: make(http.Header),
	}
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mockFlusher{},
	}

	input := `{"id":1}
`
	n, err := writer.Write([]byte(input))
	if err == nil {
		t.Fatal("Expected write error, got nil")
	}
	// Should still return the bytes we accepted from input
	if n != len(input) {
		t.Errorf("Expected to return %d bytes accepted, got %d", len(input), n)
	}
}

// failingResponseWriter always fails on Write
type failingResponseWriter struct {
	header http.Header
}

func (f *failingResponseWriter) Header() http.Header {
	return f.header
}

func (f *failingResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("mock write error")
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
}

func TestSSEResponseWriter_BufferPersistence(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Write partial message
	input1 := `{"par`
	writer.Write([]byte(input1))

	// Check buffer contains partial data
	if writer.buffer.Len() != len(input1) {
		t.Errorf("Expected buffer to contain %d bytes, got %d", len(input1), writer.buffer.Len())
	}

	// Write more partial data
	input2 := `tial":`
	writer.Write([]byte(input2))

	// Buffer should have grown
	expectedLen := len(input1) + len(input2)
	if writer.buffer.Len() != expectedLen {
		t.Errorf("Expected buffer to contain %d bytes, got %d", expectedLen, writer.buffer.Len())
	}

	// Complete the message
	input3 := `true}` + "\n"
	writer.Write([]byte(input3))

	// Buffer should be empty (or contain only remaining data)
	if writer.buffer.Len() != 0 {
		t.Errorf("Expected buffer to be empty after complete message, got %d bytes", writer.buffer.Len())
	}

	// Verify complete message was output
	output := mock.body.String()
	expected := `data: {"partial":true}` + "\n\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

func TestSSEResponseWriter_EmptyWrite(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Write empty data
	n, err := writer.Write([]byte{})
	if err != nil {
		t.Fatalf("Write empty failed: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to write 0 bytes, got %d", n)
	}

	// Should produce no output
	if mock.body.Len() > 0 {
		t.Errorf("Expected no output for empty write, got: %s", mock.body.String())
	}
}

func TestSSEResponseWriter_NilFlusher(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        nil, // nil flusher
	}

	// Write should fail with nil flusher
	_, err := writer.Write([]byte(`{"test":true}` + "\n"))
	if err == nil {
		t.Fatal("Expected error with nil flusher, got nil")
	}
	if !strings.Contains(err.Error(), "flushable") {
		t.Errorf("Expected 'flushable' error, got: %v", err)
	}
}

func TestSSEResponseWriter_ConcurrentWrites(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        &mock.mockFlusher,
	}

	// Perform concurrent writes
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			input := fmt.Sprintf(`{"id":%d}`, id) + "\n"
			_, err := writer.Write([]byte(input))
			if err != nil {
				t.Errorf("Concurrent write %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Should have written all messages (order may vary due to concurrency)
	output := mock.body.String()
	for i := 0; i < numGoroutines; i++ {
		expected := fmt.Sprintf(`"id":%d`, i)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %s", expected)
		}
	}
}

func TestSSEResponseWriter_FlushWithNilFlusher(t *testing.T) {
	mock := newMockResponseWriter()
	writer := &sseResponseWriter{
		ResponseWriter: mock,
		flusher:        nil,
	}

	// Flush should not panic with nil flusher
	writer.Flush()
}

func TestSSEResponseWriter_RealWorldScenario(t *testing.T) {
	// Simulate real httptest.ResponseRecorder behavior
	recorder := httptest.NewRecorder()
	writer := &sseResponseWriter{
		ResponseWriter: recorder,
		flusher:        recorder,
	}

	// Simulate json.Encoder writing large objects in chunks
	// This is the actual problematic scenario described in the issue
	largeJSON := `{"id":1,"validators":[` +
		strings.Repeat(`{"address":"0x1234567890abcdef","stake":1000000},`, 100) +
		`{"address":"0xfedcba0987654321","stake":1000000}]}`

	// Simulate chunked writes (what json.Encoder does)
	chunkSize := 100
	for i := 0; i < len(largeJSON); i += chunkSize {
		end := i + chunkSize
		if end > len(largeJSON) {
			end = len(largeJSON)
		}
		chunk := largeJSON[i:end]

		// Don't add newline until the end
		if end == len(largeJSON) {
			chunk += "\n"
		}

		n, err := writer.Write([]byte(chunk))
		if err != nil {
			t.Fatalf("Write chunk failed: %v", err)
		}
		if n != len(chunk) {
			t.Errorf("Expected to write %d bytes, got %d", len(chunk), n)
		}
	}

	// Verify the complete JSON was output as a single SSE event
	output := recorder.Body.String()
	if !strings.HasPrefix(output, "data: ") {
		t.Errorf("Expected output to start with 'data: ', got: %s", output[:20])
	}
	if !strings.HasSuffix(output, "\n\n") {
		t.Errorf("Expected output to end with \\n\\n, got: %s", output[len(output)-10:])
	}

	// Count number of SSE events (should be 1)
	eventCount := strings.Count(output, "data: ")
	if eventCount != 1 {
		t.Errorf("Expected 1 SSE event, got %d", eventCount)
	}

	// Verify the JSON is complete and valid
	if !strings.Contains(output, largeJSON) {
		t.Errorf("Expected output to contain complete JSON")
	}
}
