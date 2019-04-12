# hldsinfo

Go package for querying HLDS server information.

## Example

```go
info, err := hldsinfo.Get("127.0.0.1:27015", time.Time{}) // no deadline
if err != nil {
    ...
}
fmt.Println(info.Name) // prints server name
```

For example of `Fetcher` see cmd/main.go.
