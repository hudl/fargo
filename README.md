# fargo

Netflix Eureka client in golang. Named for the show Eureka.

```go
c = fargo.NewConn("http", "127.0.0.1", "8080")
c.GetApps() // returns a map[String]fargo.Application
```
