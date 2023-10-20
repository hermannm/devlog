// Package devlog implements a structured logging (slog) handler, with a human-readable output
// format designed for development builds.
//
// A devlog.Handler can be configured as follows:
//
//	logger := slog.New(devlog.NewHandler(os.Stdout, nil))
//	slog.SetDefault(logger)
//
// Following calls to slog's logging functions will use this handler, giving output on the following
// format:
//
//	slog.Info("Server started", slog.Int("port", 8000), slog.String("environment", "DEV"))
//	// [2023-10-21 10:31:09] INFO: Server started
//	// - port: 8000
//	// - environment: DEV
//
// To see the output format with colors, check the [README].
//
// [README]: https://github.com/hermannm/devlog#readme
package devlog
