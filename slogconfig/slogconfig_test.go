package slogconfig_test

import (
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"hermannm.dev/devlog/slogconfig"
)

func TestInitDefaultLogHandler(t *testing.T) {
	original := slog.Default()
	defer func() {
		slog.SetDefault(original)
	}()

	slogconfig.InitDefaultLogHandler(slog.NewTextHandler(os.Stdout, nil))

	assertWrappedHandlers(
		t,
		slog.Default().Handler(),
		[...]string{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
		[...]string{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
		[...]string{"log/slog", "TextHandler"},
	)
}

func TestInitPrettyLogHandler(t *testing.T) {
	original := slog.Default()
	defer func() {
		slog.SetDefault(original)
	}()

	slogconfig.InitPrettyLogHandler(os.Stdout, nil)

	assertWrappedHandlers(
		t,
		slog.Default().Handler(),
		[...]string{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
		[...]string{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
		[...]string{"hermannm.dev/devlog", "Handler"},
	)
}

func TestInitJSONLogHandler(t *testing.T) {
	original := slog.Default()
	defer func() {
		slog.SetDefault(original)
	}()

	slogconfig.InitJSONLogHandler(os.Stdout, nil)

	assertWrappedHandlers(
		t,
		slog.Default().Handler(),
		[...]string{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
		[...]string{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
		[...]string{"log/slog", "JSONHandler"},
	)
}

// Asserts that the given handler consists of all the given wrapped handlers.
func assertWrappedHandlers(
	t *testing.T,
	handler slog.Handler,
	expectedHandlerPkgAndType ...[2]string,
) {
	t.Helper()

	handlerValue := reflect.ValueOf(handler)
	for i, expected := range expectedHandlerPkgAndType {
		expectedPkg, expectedType := expected[0], expected[1]

		handlerType := handlerValue.Type()
		require.Equal(t, expectedType, handlerType.Name())
		require.Equal(t, expectedPkg, handlerType.PkgPath())

		// If we're not at the last handler: Unwrap inner handler. Both ctxlog and errlog keep their
		// wrapped handler as the first field, so we can check it with Field(0).
		// Interfaces and pointers must be further unwrapped with Elem().
		if i != len(expectedHandlerPkgAndType)-1 {
			handlerValue = handlerValue.Field(0)
			if handlerValue.Kind() == reflect.Interface {
				handlerValue = handlerValue.Elem()
			}
			if handlerValue.Kind() == reflect.Pointer {
				handlerValue = handlerValue.Elem()
			}
		}
	}
}
