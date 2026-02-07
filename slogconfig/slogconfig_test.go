package slogconfig_test

import (
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
		[]packageAndType{
			{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
			{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
			{"log/slog", "TextHandler"},
		},
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
		[]packageAndType{
			{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
			{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
			{"hermannm.dev/devlog", "Handler"},
		},
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
		[]packageAndType{
			{"hermannm.dev/devlog/errlog", "errorAttrHandler"},
			{"hermannm.dev/devlog/ctxlog", "contextAttrHandler"},
			{"log/slog", "JSONHandler"},
		},
	)
}

func TestNilDefaultHandler(t *testing.T) {
	assert.PanicsWithValue(
		t,
		"nil slog.Handler given to InitDefaultLogHandler",
		func() {
			slogconfig.InitDefaultLogHandler(nil)
		},
	)
}

type packageAndType struct {
	pkg      string
	typeName string
}

// Asserts that the given handler consists of all the given wrapped handlers.
func assertWrappedHandlers(
	t *testing.T,
	handler slog.Handler,
	expectedHandlerPkgsAndTypes []packageAndType,
) {
	t.Helper()

	handlerValue := reflect.ValueOf(handler)
	for i, expected := range expectedHandlerPkgsAndTypes {
		handlerType := handlerValue.Type()
		require.Equal(t, expected.pkg, handlerType.PkgPath())
		require.Equal(t, expected.typeName, handlerType.Name())

		// If we're not at the last handler: Unwrap inner handler. Both ctxlog and errlog keep their
		// wrapped handler as the first field, so we can check it with Field(0).
		// Interfaces and pointers must be further unwrapped with Elem().
		if i != len(expectedHandlerPkgsAndTypes)-1 {
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
