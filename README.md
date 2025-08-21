# devlog

A structured log handler for Go, with a human-readable output format designed for development
builds.

Run `go get hermannm.dev/devlog` to add it to your project!

## Usage

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
structured format for production, such as JSON, to make log analysis easier. You can get both by
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

## devlog/log

To complement `devlog`'s output handling, the
[`devlog/log`](https://pkg.go.dev/hermannm.dev/devlog/log) subpackage provides input handling. It is
a thin wrapper over the `slog` package, providing:

- Utility functions for log message formatting (`Infof`, `Errorf` etc.)
- Error-aware logging functions, which structure errors to be formatted nicely as log attributes
- `log.AddContextAttrs`, a function for adding log attributes to a `context.Context`, applying the
  attributes to all logs made in that context

Example using `devlog/log`:

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

This gives the following output (when using `devlog.NewHandler`):

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/ac5ebe0a372e745c30b5afe6eeb71a67c4c44d21/devlog-example-output-2.png?raw=true)

## Credits

- [Jonathan Amsterdam](https://github.com/jba) for his fantastic
  [Guide to Writing
  `slog` Handlers](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
