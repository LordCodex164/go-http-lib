# http-lib (Go)

Small HTTP helper library that provides a `RequestBuilder` wrapper around `net/http` for:

- Building requests with headers and optional JSON bodies
- Basic client-side rate limiting (RPS + burst) via `golang.org/x/time/rate`
- Simple async requests (returns a channel)

## Install

```bash
go get github.com/lordcodex164/http-lib@latest
```

## Import

The module path is `github.com/lordcodex164/http-lib`, and the package name is `httplibrary`:

```go
import httplibrary "github.com/lordcodex164/http-lib"
```

## Quick start (GET)

```go
rb := httplibrary.NewRequestBuilder("GET", "https://jsonplaceholder.typicode.com/posts/1").
  WithHeader("Authorization", "Bearer <token>")

res, err := rb.Send()
if err != nil {
  // handle error
}
defer res.Body.Close()

var out httplibrary.ApiResponse
if err := rb.ReadJsonBody(res, &out); err != nil {
  // handle error
}
```

## JSON POST + rate limiting

```go
rb := httplibrary.NewRequestBuilder("POST", "https://jsonplaceholder.typicode.com/posts").
  WithRateLimter(5, 3) // 5 req/s with burst=3

payload := httplibrary.ApiRequest{
  Title:  "Go HTTP Client",
  UserId: 23,
  Body:   "Sending data with http",
}

if err := rb.WithJsonData(payload); err != nil {
  // handle error
}

res, err := rb.Send()
if err != nil {
  // handle error
}
defer res.Body.Close()
```

## Async request

`SendAsync()` starts a goroutine, sends the request, unmarshals into an `ApiResponse`, then returns the value via a channel.

```go
rb := httplibrary.NewRequestBuilder("GET", "https://jsonplaceholder.typicode.com/posts/1")
ch := rb.SendAsync()
out := <-ch
_ = out
```

## API (current)

- `NewRequestBuilder(method, url string) *RequestBuilder`
- `(*RequestBuilder) WithHeader(key, value string) *RequestBuilder`
- `(*RequestBuilder) WithBody(body io.Reader)` (not chainable)
- `(*RequestBuilder) WithJsonData(data any) error`
- `(*RequestBuilder) WithRateLimter(rps, burst int) *RequestBuilder`
- `(*RequestBuilder) Send() (*http.Response, error)`
- `(*RequestBuilder) SendAsync() <-chan ApiResponse`
- `(*RequestBuilder) ReadJsonBody(res *http.Response, target *ApiResponse) error`
- `ExponentialBackoff(retry int) time.Duration`

## Notes / limitations

- `WithTimeout(...)` currently sets a field but is not applied to the underlying `http.Client` in `Send()`.
- Retries/backoff are not wired into `Send()` yet (only `ExponentialBackoff(...)` exists).
- Middleware support is not exposed yet (the `middlewares` field exists, but there’s no public method to register middleware).
- `ReadJsonBody(...)` currently only unmarshals into `*ApiResponse` (not an arbitrary struct).
- `WithJsonData(...)` sets header key `Content` (not `Content-Type`); you may want to also set `Content-Type: application/json` manually via `WithHeader(...)`.

## Development

```bash
go test ./... -v
```
