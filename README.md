# Rotary Phone

A Listener and Client binder using [net.Pipe](https://pkg.go.dev/net#Pipe).

## Usage:

```go
// Setup Rotary Phone Binder
conn := rotaryphone.New()
defer conn.Close()

// Setup Server bound to the Rotary Phone
wg := &sync.WaitGroup{}
wg.Add(1)
server := &http.Server{Handler: setupMux()}
go func() {
    defer wg.Done()
    _ = server.Serve(conn)
}()
defer wg.Wait()
defer server.Close()

// Create a new HTTP client
client := conn.Client()

// Form and execute request
req := formRequest()
client.Do(req)
```
