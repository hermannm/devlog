// Package devlog implements a structured log (slog) handler, with a human-readable output format
// designed for local development and CLI tools.
//
// A devlog.Handler can be configured as follows:
//
//	logHandler := devlog.NewHandler(os.Stdout, nil)
//	slog.SetDefault(slog.New(logHandler))
//
// Following calls to [log/slog]'s logging functions will use this handler, giving output on the
// following format:
//
//	slog.Info("Server started", "port", 8000, "environment", "DEV")
//	// [10:31:09] INFO: Server started
//	//   port: 8000
//	//   environment: DEV
//
// Check the [README] to see the output format with colors.
//
// You can also use the [hermannm.dev/devlog/sloginit] package for shorter initialization:
//
//	sloginit.InitPrettyLogHandler(os.Stdout, nil)
//
// ...which also configures [hermannm.dev/devlog/errlog.NewHandler] and
// [hermannm.dev/devlog/ctxlog.NewHandler].
package devlog
