# Symbiotic Relay Client Examples

This directory contains example code demonstrating how to use the [Symbiotic Relay Go client library](../v1/) to interact with a Symbiotic Relay server.

## Basic Usage Example

The example shows how to:

1. Connect to a Symbiotic Relay server
2. Get current epoch information
3. Sign messages 
4. Retrieve aggregation proofs and signatures
5. Get validator set information
6. Use streaming responses for real-time updates

## Prerequisites

Before running the examples, ensure you have:

- **[Go 1.24 or later](https://golang.org/doc/install)** installed
- **Access to a running Symbiotic Relay Network**
- **Network connectivity** to the relay server
- **Valid key configurations** on the relay server (for signing operations)

## Running the Example

```bash
cd api/client/examples
go run main.go
```

By default, the example will try to connect to `localhost:8080`. You can specify a different server URL by setting the `RELAY_SERVER_URL` environment variable:

```bash
RELAY_SERVER_URL=my-relay-server:8081 go run main.go
```

NOTE: for the signature/proof generation to work you need to run the script for all active relay servers to get the majority consensus to generate proof.


## Integration with Your Application

To integrate this client into your own application:

1. **Import the client package**:
   ```go
   import client "github.com/symbioticfi/relay/api/client/v1"
   ```

2. **Create a connection**:
   ```go
   conn, err := grpc.Dial(serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
   if err != nil {
       return err
   }
   client := client.NewSymbioticClient(conn)
   ```

3. **Use the client methods** as demonstrated in the [example](main.go)

4. **Handle errors appropriately** for your use case

5. **Ensure proper connection cleanup** with `defer conn.Close()`

## More Examples

For a more comprehensive example of using the client library in a real-world application, see:

- **[Symbiotic Super Sum Example](https://github.com/symbioticfi/symbiotic-super-sum/tree/main/off-chain)** 

## API Reference

For complete API documentation, refer to:

- **API Documentation**: [`docs/api/v1/doc.md`](../../../docs/api/v1/doc.md)
- **Protocol Buffer definitions**: [`api/proto/v1/api.proto`](../../proto/v1/api.proto)
- **Generated Go types**: [`api/client/v1/types.go`](../v1/types.go)
- **Client interface**: [`api/client/v1/client.go`](../v1/client.go)


## License

This client library and example code are licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.