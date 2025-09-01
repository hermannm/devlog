# devlog

Go library that provides utilities for structured logging, building on the standard
[`log/slog`](https://pkg.go.dev/log/slog) package. It provides two independent packages:

- `devlog` implements a [`slog.Handler`](https://pkg.go.dev/log/slog#Handler) with a human-readable
  output format, designed for local development and CLI tools
- `devlog/log` is a thin wrapper over the logging API of `log/slog`, providing:
    - Utility functions for log message formatting (`log.Infof`, `log.Errorf`, etc.)
    - Error-aware logging functions, which structure errors to be formatted consistently as log
      attributes
    - `log.AddContextAttrs`, a function for adding log attributes to a
      [`context.Context`](https://pkg.go.dev/context), applying the attributes to all logs made in
      that context

Run `go get hermannm.dev/devlog` to add it to your project!

**Docs:** [pkg.go.dev/hermannm.dev/devlog](https://pkg.go.dev/hermannm.dev/devlog)

**Contents:**

- [Usage](#usage)
    - [Using the `devlog` output handler](#using-the-devlog-output-handler)
    - [Using the `devlog/log` logging API](#using-the-devloglog-logging-api)
- [Developer's guide](#developers-guide)
- [Credits](#credits)

## Usage

### Using the `devlog` output handler

`devlog.Handler` implements [`slog.Handler`](https://pkg.go.dev/log/slog#Handler), so it can handle
output for `slog`'s logging functions. It can be configured as follows:

<!-- @formatter:off -->
```go
import (
	"log/slog"

	"hermannm.dev/devlog"
)

func main() {
	logHandler := devlog.NewHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(logHandler))
}
```
<!-- @formatter:on -->

Logging with `slog` will now use this handler. So the following log:

<!-- @formatter:off -->
```go
slog.Info("Server started", "port", 8000, "environment", "DEV")
```
<!-- @formatter:on -->

...will give the following output (using a gruvbox terminal color scheme):

![Screenshot of log message in a terminal](https://github.com/hermannm/devlog/blob/3089fbac4d2cecd3d55b422a7ba742f788d5dace/devlog-example-output.png?raw=true)

Structs, slices and other non-primitive types are encoded as pretty-formatted JSON, so this
example:

<!-- @formatter:off -->
```go
type Event struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
event := Event{ID: 1000, Type: "ORDER_UPDATED"}

slog.Error("Failed to process event", "event", event)
```
<!-- @formatter:on -->

...gives this output:

![Screenshot of log message in a terminal](https://github.com/hermannm/devlog/blob/3089fbac4d2cecd3d55b422a7ba742f788d5dace/devlog-example-output-2.png?raw=true)

`devlog`'s output is meant to be easily read by a developer working locally. However, you may want a
more structured format for production systems, to make log analysis easier. You can get both by
conditionally choosing the log handler for your application, like this:

<!-- @formatter:off -->
```go
var logHandler slog.Handler
switch os.Getenv("ENVIRONMENT") {
case "LOCAL", "TEST":
	// Pretty-formatted logs for local development and tests
	logHandler = devlog.NewHandler(os.Stdout, nil)
default:
	// Structured JSON logs for deployed environments
	logHandler = slog.NewJSONHandler(os.Stdout, nil)
}
slog.SetDefault(slog.New(logHandler))
```
<!-- @formatter:on -->

### Using the `devlog/log` logging API

Unlike `log/slog`, `devlog/log` provides logging functions that take an `error`. When an error is
passed to such a logging function, it is attached to the log as a `cause` attribute, so errors are
structured consistently between logs.

<!-- @formatter:off -->
```go
import (
	"context"
	"errors"

	"hermannm.dev/devlog/log"
)

func example(ctx context.Context) {
	err := errors.New("database insert failed")
	log.Error(ctx, err, "Failed to store event")
}
```
<!-- @formatter:on -->

This gives the following output (using the `devlog` output handler):

![Screenshot of log message in a terminal](https://github.com/hermannm/devlog/blob/3089fbac4d2cecd3d55b422a7ba742f788d5dace/devlog-example-output-3.png?raw=true)

The package also provides `log.AddContextAttrs`, a function for adding log attributes to a
`context.Context`. These attributes are added to all logs where the context is passed, so this
example:

<!-- @formatter:off -->
```go
func processEvent(ctx context.Context, event Event) {
	ctx = log.AddContextAttrs(ctx, "eventId", event.ID)

	log.Debug(ctx, "Started processing event")
	// ...
	log.Debug(ctx, "Finished processing event")
}
```
<!-- @formatter:on -->

...gives this output:

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/3089fbac4d2cecd3d55b422a7ba742f788d5dace/devlog-example-output-4.png?raw=true)

This can help you trace connected logs in your system (especially when using a more structured JSON
output in production, allowing you to filter on all logs with a specific `eventId`).

In order to encourage propagating context attributes, all log functions in this package take a
`context.Context`. If you're in a function without a context parameter, you may pass a `nil`
context. But ideally, you should pass a context wherever you do logging, in order to propagate
context attributes.

## Developer's guide

When publishing a new release:

- Run tests and linter ([`golangci-lint`](https://golangci-lint.run/)):
  ```
  go test ./... && golangci-lint run
  ```
- Add an entry to `CHANGELOG.md` (with the current date)
    - Remember to update the link section, and bump the version for the `[Unreleased]` link
- Create commit and tag for the release (update `TAG` variable in below command):
  ```
  TAG=vX.Y.Z && git commit -m "Release ${TAG}" && git tag -a "${TAG}" -m "Release ${TAG}" && git log --oneline -2
  ```
- Push the commit and tag:
  ```
  git push && git push --tags
  ```
    - Our release workflow will then create a GitHub release with the pushed tag's changelog entry

## Credits

- [Jonathan Amsterdam](https://github.com/jba) for his fantastic
  [Guide to Writing
  `slog` Handlers](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
