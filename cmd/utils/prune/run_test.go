package prune

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectStorageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     []string
		want      string
		wantError bool
	}{
		{
			name:  "bbolt only — relay.db",
			files: []string{"relay.db"},
			want:  storageTypeBbolt,
		},
		{
			name:  "badger only — MANIFEST",
			files: []string{"MANIFEST"},
			want:  storageTypeBadger,
		},
		{
			name:  "badger only — *.vlog",
			files: []string{"000001.vlog"},
			want:  storageTypeBadger,
		},
		{
			name:      "both present — ambiguous",
			files:     []string{"relay.db", "MANIFEST"},
			wantError: true,
		},
		{
			name:      "neither present — empty dir",
			files:     nil,
			wantError: true,
		},
		{
			name:      "unrelated file — neither match",
			files:     []string{"README.md"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, name := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600))
			}

			got, err := detectStorageType(dir)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDetectStorageType_NonexistentDir(t *testing.T) {
	t.Parallel()
	_, err := detectStorageType(filepath.Join(t.TempDir(), "does-not-exist"))
	require.Error(t, err)
}

func TestHumanBytes_ClampedAtExabyte(t *testing.T) {
	t.Parallel()
	// math.MaxInt64 ≈ 8 EiB, the largest int64 value — must clamp to E (exa)
	// without indexing past the suffix table.
	require.NotPanics(t, func() {
		_ = humanBytes(1 << 60)       // 1 EiB
		_ = humanBytes(math.MaxInt64) // ~8 EiB
	})
}
