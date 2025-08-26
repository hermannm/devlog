# Changelog

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

[v0.5.0]: https://github.com/hermannm/devlog/compare/v0.4.1...v0.5.0

[v0.4.1]: https://github.com/hermannm/devlog/compare/v0.4.0...v0.4.1

[v0.4.0]: https://github.com/hermannm/devlog/compare/v0.3.2...v0.4.0

[v0.3.2]: https://github.com/hermannm/devlog/compare/v0.3.1...v0.3.2

[v0.3.1]: https://github.com/hermannm/devlog/compare/v0.3.0...v0.3.1

[v0.3.0]: https://github.com/hermannm/devlog/compare/v0.2.0...v0.3.0

[v0.2.0]: https://github.com/hermannm/devlog/compare/v0.1.0...v0.2.0

[v0.1.0]: https://github.com/hermannm/devlog/compare/3981051...v0.1.0
