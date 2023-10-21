# devlog

A structured logging handler for Go, with a human-readable output format designed for development
builds.

Run `go get hermannm.dev/devlog` to add it to your project!

## Usage

`devlog.Handler` implements [`slog.Handler`](https://pkg.go.dev/log/slog#Handler), so it can handle
output for `slog`'s logging functions. It can be configured as follows:

```go
logger := slog.New(devlog.NewHandler(os.Stdout, nil))
slog.SetDefault(logger)
```

Logging with `slog` will now use this handler:

```go
slog.Warn("No value found for 'PORT' in env, defaulting to 8000")
slog.Info("Server started", slog.Int("port", 8000), slog.String("environment", "DEV"))
slog.Error(
	"Database query failed",
	slog.Group("dbError", slog.Int("code", 60), slog.String("message", "UNKNOWN_TABLE")),
)
```

...giving the following output (using a gruvbox terminal color scheme):

![Screenshot of 3 log messages in a terminal](https://github.com/hermannm/devlog/blob/assets/devlog-example-output.png?raw=true)

This output is meant to be easily read by a developer working locally. However, you may want a more
structured format for production, such as JSON, to make log analysis easier. You can get both by
conditionally choosing the log handler for your application, e.g.:

```go
var handler slog.Handler
switch os.Getenv("ENVIRONMENT") {
case "PROD":
	handler = slog.NewJSONHandler(os.Stdout, nil)
case "DEV":
	handler = devlog.NewHandler(os.Stdout, nil)
}

slog.SetDefault(slog.New(handler))
```

## Credits

- [Jonathan Amsterdam](https://github.com/jba) for his fantastic [Guide to Writing `slog` Handlers](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
