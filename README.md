# devlog

A structured logging handler for Go, with a human-readable output format designed for development
builds.

Run `go get hermannm.dev/devlog` to add it to your project!

## Usage

`devlog.Handler` implements [`slog.Handler`](https://pkg.go.dev/log/slog#Handler), so it can handle
output for `slog`'s logging functions. It can be configured as follows:

```go
logHandler := devlog.NewHandler(os.Stdout, nil)
slog.SetDefault(slog.New(logHandler))
```

Logging with `slog` will now use this handler:

```go
slog.Warn("no value found for 'PORT' in env, defaulting to 8000")
slog.Info("server started", slog.Int("port", 8000), slog.String("environment", "DEV"))
slog.Error(
	"database query failed",
	slog.Group("dbError", slog.Int("code", 60), slog.String("message", "UNKNOWN_TABLE")),
)
```

...giving the following output (using a gruvbox terminal color scheme):

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/ac14f0dc1823316c983fb9cef6f1cf73a0bbb923/devlog-example-output.png?raw=true)

This output is meant to be easily read by a developer working locally. However, you may want a more
structured format for production, such as JSON, to make log analysis easier. You can get both by
conditionally choosing the log handler for your application, e.g.:

```go
var logHandler slog.Handler
switch os.Getenv("ENVIRONMENT") {
case "PROD":
	logHandler = slog.NewJSONHandler(os.Stdout, nil)
case "DEV":
	logHandler = devlog.NewHandler(os.Stdout, nil)
}

slog.SetDefault(slog.New(logHandler))
```

## devlog/log

To complement `devlog`'s output handling, the
[`devlog/log`](https://pkg.go.dev/hermannm.dev/devlog/log) subpackage provides input handling. It is
a thin wrapper over the `slog` package, with utility functions for log message formatting.

Example using `devlog` and `devlog/log` together:

```go
import (
	"errors"
	"log/slog"
	"os"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
)

func main() {
	logHandler := devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logHandler))

	err := errors.New("invalid username")
	log.ErrorCause(err, "failed to create user") // Uses slog.Default()

	userJSON := map[string]any{"id": 1, "username": "hermannm"}
	log.DebugJSON(userJSON, "user")
}
```

This gives the following output:

![Screenshot of log messages in a terminal](https://github.com/hermannm/devlog/blob/ac14f0dc1823316c983fb9cef6f1cf73a0bbb923/devlog-example-output-2.png?raw=true)

## Credits

- [Jonathan Amsterdam](https://github.com/jba) for his fantastic
  [Guide to Writing `slog` Handlers](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
