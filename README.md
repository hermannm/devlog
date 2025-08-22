# devlog

Go library that provides utilities for structured logging, building on the standard
[`log/slog`](https://pkg.go.dev/log/slog) package. It provides two independent packages:

- `devlog` implements a [`slog.Handler`](https://pkg.go.dev/log/slog#Handler) with a human-readable
  output format, designed for local development and CLI tools
- `devlog/log` is a thin wrapper over the logging API of `log/slog`, providing:
    - Utility functions for log message formatting (`log.Infof`, `log.Errorf`, etc.)
    - Error-aware logging functions, which structure errors to be formatted nicely as log attributes
    - `log.AddContextAttrs`, a function for adding log attributes to a
      [`context.Context`](https://pkg.go.dev/context), applying the attributes to all logs made in
      that context

Run `go get hermannm.dev/devlog` to add it to your project!

**Docs:** [pkg.go.dev/hermannm.dev/devlog](https://pkg.go.dev/hermannm.dev/devlog)

**Contents:**

- [Usage](#usage)
    - [Using the `devlog` output handler](#using-the-devlog-output-handler)
    - [Using the `devlog/log` logging API](#using-the-devloglog-logging-api)
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

Logging with `slog` will now use this handler:

<!-- @formatter:off -->
```go
slog.Warn("No value found for 'PORT' in env, defaulting to 8000")
slog.Info("Server started", "port", 8000, "environment", "DEV")
slog.Error(
	"Database query failed",
	slog.Group("dbError", "code", 60, "message", "UNKNOWN_TABLE"),
)
```
<!-- @formatter:on -->

...giving the following output (using a gruvbox terminal color scheme):

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/ac5ebe0a372e745c30b5afe6eeb71a67c4c44d21/devlog-example-output.png?raw=true)

This output is meant to be easily read by a developer working locally. However, you may want a more
structured format (like JSON) for production systems, to make log analysis easier. You can get both
by conditionally choosing the log handler for your application, like this:

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

Example:

<!-- @formatter:off -->
```go
import (
	"context"
	"errors"

	"hermannm.dev/devlog/log"
)

func example(ctx context.Context) {
	user := map[string]any{"id": 2, "username": "hermannm"}
	err := errors.New("username taken")
	log.Error(ctx, err, "Failed to create user", "user", user)
}
```
<!-- @formatter:on -->

This gives the following output (using the `devlog` output handler):

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/ac5ebe0a372e745c30b5afe6eeb71a67c4c44d21/devlog-example-output-2.png?raw=true)

## Credits

- [Jonathan Amsterdam](https://github.com/jba) for his fantastic
  [Guide to Writing
  `slog` Handlers](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
