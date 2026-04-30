# devlog

Go library that provides utilities for structured logging, building on the standard
[`log/slog`](https://pkg.go.dev/log/slog) package. It consists of the following packages:

- `devlog` provides a [`slog.Handler`](https://pkg.go.dev/log/slog#Handler) with a human-readable
  output format, designed for local development and CLI tools.
- `errlog` provides a wrapping `slog.Handler` which transforms error log attributes to make them
  more structured.
- `ctxlog` provides a way to attach log attributes to a
  [`context.Context`](https://pkg.go.dev/context), so that all logs in the scope of that context
  get those attributes in their output.
- `sloginit` provides utility functions for initializing the default `slog` handler with `errlog`'s
  and `ctxlog`'s wrapping handlers applied.

Run `go get hermannm.dev/devlog` to add it to your project!

**Docs:** [pkg.go.dev/hermannm.dev/devlog](https://pkg.go.dev/hermannm.dev/devlog)

**Contents:**

- [Usage](#usage)
    - [Using the `devlog` output handler](#using-the-devlog-output-handler)
    - [Using the `devlog/log` logging API](#using-the-devloglog-logging-api)
- [Maintainer's guide](#maintainers-guide)
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
	"hermannm.dev/devlog/sloginit"
)

func main() {
	slog.SetDefault(slog.New(devlog.NewHandler(os.Stdout, nil)))

	// Alternatively, use the sloginit package:
	sloginit.InitPrettyLogHandler(os.Stdout, nil)
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
switch os.Getenv("ENVIRONMENT") {
case "LOCAL", "TEST":
	// Pretty-formatted logs for local development and tests
	sloginit.InitPrettyLogHandler(os.Stdout, nil)
default:
	// Structured JSON logs for deployed environments
	sloginit.InitJSONLogHandler(os.Stdout, nil)
}
```
<!-- @formatter:on -->

### Using `errlog` for structured error attributes

`errlog` transforms `slog.Attr`s with `error` values to give them more structure. You use it by
wrapping your `slog.Handler` with `errlog.NewHandler`:

<!-- @formatter:off -->
```go
import (
	"log/slog"
	"os"

	"hermannm.dev/devlog/errlog"
	"hermannm.dev/devlog/sloginit"
)

func main() {
	slog.SetDefault(
		slog.New(
			errlog.NewHandler(
				slog.NewJSONHandler(os.Stdout, nil),
			),
		),
	)

	// Alternatively, use sloginit, which applies errlog.NewHandler for you:
	sloginit.InitJSONLogHandler(os.Stdout, nil)
}
```
<!-- @formatter:on -->

Now, when you log errors with `slog` like this:

<!-- @formatter:off -->
```go
import (
	"errors"
	"fmt"
	"log/slog"
)

func main() {
	if err := fallibleFunction(); err != nil {
		slog.Error("Something went wrong", "error", err)
	}
}

func fallibleFunction() {
	if err := innerFunction(); err != nil {
		return fmt.Errorf("inner function failed: %w", err)
	}
}

func innerFunction() {
	return errors.New("root cause")
}
```
<!-- @formatter:on -->

...you get the following JSON output:

<!-- @formatter:off -->
```json
{
  "time": "...",
  "level": "ERROR",
  "msg": "Something went wrong",
  "error": {
    "msg": "inner function failed",
    "cause": {
      "msg": "root cause"
    }
  }
}
```
<!-- @formatter:on -->

### Using `ctxlog` for context attributes

`ctxlog` provides `ctxlog.WithAttrs`, a function for adding log attributes to a
[`context.Context`](https://pkg.go.dev/context). In order to use this, you must first wrap your
`slog.Handler` with `ctxlog.NewHandler`:

<!-- @formatter:off -->
```go
import (
	"log/slog"
	"os"

	"hermannm.dev/devlog/ctxlog"
	"hermannm.dev/devlog/sloginit"
)

func main() {
	slog.SetDefault(
		slog.New(
			ctxlog.NewHandler(
				slog.NewJSONHandler(os.Stdout, nil),
			),
		),
	)

	// Alternatively, use sloginit, which applies ctxlog.NewHandler for you:
	sloginit.InitJSONLogHandler(os.Stdout, nil)
}
```
<!-- @formatter:on -->

Now you can add context attributes with `ctxlog.WithAttrs`:

<!-- @formatter:off -->
```go
func processEvent(ctx context.Context, event Event) {
	ctx = ctxlog.WithAttrs(ctx, "eventId", event.ID)

	slog.InfoContext(ctx, "Processing event")
	// ...
	slog.InfoContext(ctx, "Successfully processed event")
}
```
<!-- @formatter:on -->

...giving this output:

<!-- @formatter:off -->
```json lines
{ "time":"...", "level": "INFO", "msg": "Processing event", "eventId": 1000 }
{ "time":"...", "level": "INFO", "msg": "Successfully processed event", "eventId": 1000 }
```
<!-- @formatter:on -->

This can help you trace all logs in the scope of this event's processing, by filtering on the
`eventId` in your log analysis tool. You may want to use [OpenTelemetry](https://opentelemetry.io/)
for more comprehensive tracing, but `ctxlog` lets you get basic tracing with just the standard
`slog` package.

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

In order to encourage propagating context attributes, all log functions in this package take a
`context.Context`. If you're in a function without a context parameter, you may pass a `nil`
context. But ideally, you should pass a context wherever you do logging, in order to propagate
context attributes.

## Maintainer's guide

### Publishing a new release

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
