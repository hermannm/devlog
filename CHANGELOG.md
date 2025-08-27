# Changelog

## [v0.6.0] - 2025-08-27

- `devlog`:
    - **Breaking:** Remove `InitDefaultLogHandler`
        - This function made it harder to configure other log handlers, which we do not want
        - To migrate, replace:
          ```go
          devlog.InitDefaultLogHandler(output, options)
          ```
          ...with:
          ```go
          logHandler := devlog.NewHandler(output, options)
          slog.SetDefault(slog.New(logHandler))
          ```
            - Alternatively, use the new `log.SetDefault` function from `devlog/log` (see notes
              below):
              ```go
              log.SetDefault(devlog.NewHandler(output, options))
              ```
    - Encode log attributes with values of type `any` as pretty-formatted JSON
        - This makes `devlog`'s attribute encoding more consistent with the JSON log handler that
          you'll typically use for production, to avoid discrepancies between local and production
          log output
- `devlog/log`:
    - Add `log.AddContextAttrs` function for attaching log attributes to a `context.Context`
    - **Breaking:** Add `context.Context` parameter to all logging functions, to ensure that
      attributes from `AddContextAttrs` are propagated
        - To migrate, replace:
          ```go
          log.Info("Message", ...attributes)
          ```
          ...with:
          ```go
          log.Info(ctx, "Message", ...attributes)
          ```
            - If you're in a function without a context parameter, you may pass a `nil` context:
              ```go
              log.Info(nil, "Message", ...attributes)
              ```
    - **Breaking:** Remove `log.Error`, `log.WarnError` and `log.DebugError` functions that don't
      take a log message, and replace them with the error-logging functions that previously had a
      `Cause` suffix
        - Renamed functions:
            - `log.ErrorCause` -> `log.Error`
            - `log.ErrorCausef` -> `log.Errorf`
            - `log.WarnErrorCause` -> `log.WarnError`
            - `log.WarnErrorCausef` -> `log.WarnErrorf`
            - `log.DebugErrorCause` -> `log.DebugError`
            - `log.DebugErrorCausef` -> `log.DebugErrorf`
        - To migrate uses of `Error`, `WarnError` and `DebugError` without a log message, you can
          add a blank log message, which behaves the same (error string is used as the message).
          Replace:
          ```go
          log.Error(err, ...attributes)
          log.WarnError(err, ...attributes)
          log.DebugError(err, ...attributes)
          ```
          ...with:
          ```go
          log.Error(ctx, err, "", ...attributes)
          log.WarnError(ctx, err, "", ...attributes)
          log.DebugError(ctx, err, "", ...attributes)
          ```
    - **Breaking:** Change variadic parameter in `log.Errors` to slice of errors, to allow passing
      log attributes
        - To migrate, replace:
          ```go
          log.Errors("Message", err1, err2)
          ```
          ...with:
          ```go
          log.Errors(ctx, []error{err1, err2}, "Message")
          ```
    - Add functions for wrapping multiple errors with a formatted message:
        - `log.Errorsf`
        - `log.WarnErrorsf`
        - `log.DebugErrorsf`
    - Add functions for logging errors at the `INFO` log level (for completeness with the other
      error-logging functions):
        - `log.InfoError`
        - `log.InfoErrorf`
        - `log.InfoErrors`
        - `log.InfoErrorsf`
    - Add log functions that allow passing the log level as a parameter, to dynamically set the
      level
        - `log.Log`
        - `log.Logf`
        - `log.LogWithError`
        - `log.LogWithErrorf`
        - `log.LogWithErrors`
        - `log.LogWithErrorsf`
    - **Breaking:** Remove `log.JSON` attribute function
        - This is redundant now that `devlog` encodes `any` values as JSON
    - Add support for errors with structured log attributes attached
        - The [hermannm.dev/wrap](https://pkg.go.dev/hermannm.dev/wrap) package uses this, in its
          `wrap.ErrorWithAttrs` function
    - Add support for errors with `context.Context` attached, in order to propagate attributes from
      `log.AddContextAttrs` from the error's original context
        - The [hermannm.dev/wrap/ctxwrap](https://pkg.go.dev/hermannm.dev/wrap/ctxwrap) package uses
          this
    - Add `log.ContextHandler`, which wraps a `slog.Handler` to support attributes from
      `log.AddContextAttrs` for logs made outside of this package
    - Add `log.SetDefault`, short-hand utility for calling
      `slog.SetDefault(slog.New(log.ContextHandler(logHandler)))`
    - Add `log.Enabled` and `Logger.Enabled` for checking if log output is enabled for a log level
    - Make error unwrapping of plain errors more robust (check for `Unwrap() error` method instead
      of just splitting on ": ")

## [v0.5.0] - 2024-10-02

- `devlog`:
    - Change time format to no longer show date by default (when in local development, one typically
      only cares about time)
    - Add `TimeFormat` option to include date in time format, if one still wants the old behavior
    - Add `InitDefaultLogHandler` convenience function, that combines `NewHandler` and
      `slog.SetDefault`
    - Add function name when `AddSource` is enabled, and fix formatting of `source` attribute
    - Improve handling of `Handler.WithAttrs`
        - Logs now display log-specific attributes first (i.e. most recent first), and attributes
          from `WithAttrs` after
    - Fix handling of `Handler.WithGroup` when not followed by `Handler.WithAttrs`
        - The previous implementation was broken, leading to subsequent logs getting increasing
          levels of indentation
- `devlog/log`:
    - Allow passing log attributes as `...any` (key/value pairs) instead of just `...slog.Attr`,
      enabling more concise attribute syntax
    - Add `DebugError`, `DebugErrorCause`, `DebugErrorCausef` and `DebugErrors` functions (both at
      top level and for `Logger`), for error logging at `DEBUG` level
    - **Breaking:** Make `JSONValue` struct no longer public (you likely were not depending on this)
    - **Breaking:** Make `WrappedError` and `WrappedErrors` interfaces no longer public (you likely
      were not depending on these)

## [v0.4.1] - 2023-12-03

- `devlog`: Fix `IsColorTerminal` for windows

## [v0.4.0] - 2023-12-03

- `devlog`:
    - Add `ForceColor` option to enable colors regardless of terminal color support
- `devlog/log`:
    - Add `JSON` log attribute function
        - Logs a value as prettified JSON when using `devlog.Handler`, or normal JSON when using
          `slog.JSONHandler`
    - **Breaking:** Remove `DebugJSON` functions, made redundant by `JSON` log attribute
    - **Breaking:** Rename `ErrorWarning` / `ErrorWarningf` -> `WarnErrorCause` / `WarnErrorCausef`
      for more consistent naming
    - Add `WarnError` function
- **Breaking:** Remove `color` package
    - Merged into `devlog`, as that's the only place it's used now

## [v0.3.2] - 2023-11-16

- `devlog/log`: Add `Default` constructor for `Logger`

## [v0.3.1] - 2023-11-16

- `devlog/log`: Change `Logger` to be passed by value

## [v0.3.0] - 2023-11-15

- `devlog`:
    - Change log attribute output format to reduce noise
    - Add formatting for list values in log attributes
- `devlog/log`:
    - Add `Logger` type to enable use of `With` and `WithGroup` for shared log attributes
    - Change `log.Error` to print only error, and add `log.ErrorCause` to log error with message
    - Add error interfaces to decouple dependency on `hermannm.dev/wrap`
    - Implement splitting of long error messages
    - Add `DebugJSON` utility function (using [
      `neilotoole/jsoncolor`](https://github.com/neilotoole/jsoncolor) for colors)
- Add `color` subpackage for shared color constants and terminal color support check

## [v0.2.0] - 2023-10-27

- Add `devlog/log` subpackage for log message formatting

## [v0.1.0] - 2023-10-21

- Initial release

[Unreleased]: https://github.com/hermannm/devlog/compare/v0.6.0...HEAD

[v0.6.0]: https://github.com/hermannm/devlog/compare/v0.5.0...v0.6.0

[v0.5.0]: https://github.com/hermannm/devlog/compare/v0.4.1...v0.5.0

[v0.4.1]: https://github.com/hermannm/devlog/compare/v0.4.0...v0.4.1

[v0.4.0]: https://github.com/hermannm/devlog/compare/v0.3.2...v0.4.0

[v0.3.2]: https://github.com/hermannm/devlog/compare/v0.3.1...v0.3.2

[v0.3.1]: https://github.com/hermannm/devlog/compare/v0.3.0...v0.3.1

[v0.3.0]: https://github.com/hermannm/devlog/compare/v0.2.0...v0.3.0

[v0.2.0]: https://github.com/hermannm/devlog/compare/v0.1.0...v0.2.0

[v0.1.0]: https://github.com/hermannm/devlog/compare/3981051...v0.1.0
