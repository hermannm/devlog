package devlog_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"hermannm.dev/devlog"
)

// Tests our handler against the standard library test suite for structured log handlers.
func TestSlog(t *testing.T) {
	var buffer bytes.Buffer

	slogtest.Run(
		t,
		func(t *testing.T) slog.Handler {
			buffer.Reset()
			return devlog.NewHandler(&buffer, &devlog.Options{
				DisableColors: true,
				TimeFormat:    devlog.TimeFormatFull,
			})
		},
		func(t *testing.T) map[string]any {
			entries, err := parseLogEntry(buffer.String())
			if err != nil {
				t.Fatal(err)
			}
			return entries
		},
	)
}

func TestTimeFormat(t *testing.T) {
	timeValue, err := time.Parse(time.DateTime, "2024-09-29 10:57:30")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		format         devlog.TimeFormat
		expectedOutput string
	}{
		{
			format:         devlog.TimeFormatShort,
			expectedOutput: "[10:57:30]",
		},
		{
			format:         devlog.TimeFormatFull,
			expectedOutput: "[2024-09-29 10:57:30]",
		},
	}

	for _, testCase := range testCases {
		var buffer bytes.Buffer
		handler := devlog.NewHandler(&buffer, &devlog.Options{
			DisableColors: true,
			TimeFormat:    testCase.format,
		})

		record := slog.NewRecord(timeValue, slog.LevelInfo, "Message", 0)
		if err := handler.Handle(context.Background(), record); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		assertContains(t, buffer.String(), testCase.expectedOutput)
	}
}

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestStructAttr(t *testing.T) {
	user := user{
		ID:   1,
		Name: "hermannm",
	}

	output := getLogOutput(t, nil, func() {
		slog.Info("user created", "user", user)
	})

	expectedOutput := `user: {
    "id": 1,
    "name": "hermannm"
  }`

	assertContains(t, output, expectedOutput)
}

func TestListAttrs(t *testing.T) {
	type testStruct struct {
		Field string `json:"field"`
	}

	testCases := []struct {
		attr           slog.Attr
		expectedOutput string
	}{
		{
			attr: slog.Any("stringList", []string{"test1", "test2", "test3"}),
			expectedOutput: `  stringList: [
    "test1",
    "test2",
    "test3"
  ]`,
		},
		{
			attr: slog.Any("structList", []testStruct{{"test1"}, {"test2"}}),
			expectedOutput: `  structList: [
    {
      "field": "test1"
    },
    {
      "field": "test2"
    }
  ]`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.attr.Key, func(t *testing.T) {
			output := getLogOutput(t, nil, func() {
				slog.Info("", testCase.attr)
			})
			assertContains(t, output, testCase.expectedOutput)
		})
	}
}

func TestCauseError(t *testing.T) {
	// This follows the structure that the devlog/log subpackage uses for logged errors
	errorLog := []any{
		"something went wrong",
		"cause error 1",
		[]any{"underlying cause 1"},
		"cause error 2",
		[]any{"underlying cause 2", "underlying cause 3"},
	}

	output := getLogOutput(t, nil, func() {
		slog.Error("", "cause", errorLog)
	})

	assertContains(
		t,
		output,
		`  cause:
    - something went wrong
    - cause error 1
      - underlying cause 1
    - cause error 2
      - underlying cause 2
      - underlying cause 3`,
	)
}

func TestSource(t *testing.T) {
	output := getLogOutput(t, &devlog.Options{AddSource: true}, func() {
		slog.Info("Test")
	})

	assertContains(
		t,
		output,
		"\n  source: hermannm.dev/devlog_test.TestSource",
		"handler_test.go:167",
	)
}

func getLogOutput(t *testing.T, handlerOptions *devlog.Options, logFunc func()) string {
	if handlerOptions == nil {
		handlerOptions = &devlog.Options{}
	}
	// Must disable color output to parse reliably
	handlerOptions.DisableColors = true

	var buffer bytes.Buffer
	slog.SetDefault(slog.New(devlog.NewHandler(&buffer, handlerOptions)))

	logFunc()

	output := buffer.String()
	t.Log(strings.TrimSuffix(output, "\n"))
	return output
}

func assertContains(t *testing.T, output string, expectedInOutput ...string) {
	t.Helper()

	output = strings.TrimSuffix(output, "\n")

	for _, expected := range expectedInOutput {
		if !strings.Contains(output, expected) {
			t.Errorf(`unexpected log output
got:
----------------------------------------
%s
----------------------------------------

want:
----------------------------------------
%s
----------------------------------------
`, output, expected)
		}
	}
}

// slogtest.Run requires us to parse our log output to a map[string]any.
func parseLogEntry(entryString string) (map[string]any, error) {
	entry := make(map[string]any)

	entryString, includeTime := strings.CutPrefix(entryString, "[")
	if includeTime {
		split := strings.SplitN(entryString, "] ", 2)
		timeString := split[0]
		entryString = split[1]

		timeValue, err := time.Parse(time.DateTime, timeString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %w", err)
		}
		entry[slog.TimeKey] = timeValue
	}

	split := strings.SplitN(entryString, ": ", 2)
	entry[slog.LevelKey] = split[0]

	// Cut trailing newline
	entryString, _ = strings.CutSuffix(split[1], "\n")

	split = strings.SplitN(entryString, "\n", 2)
	entry[slog.MessageKey] = split[0]

	hasAttributes := len(split) == 2
	if hasAttributes {
		entryString = split[1]

		var openGroups []string
		currentIndent := 0
		for _, line := range strings.Split(entryString, "\n") {
			attr := strings.TrimLeft(line, " ")
			indent := (len(line) - len(attr) - 1) / 2

			if indent < currentIndent {
				openGroups = openGroups[0 : len(openGroups)-1]
				currentIndent--
			}

			split = strings.SplitN(attr, ":", 2)
			attrKey := split[0]
			attrValue := split[1]

			subEntry := getSubEntry(entry, openGroups)

			if attrValue == "" {
				openGroups = append(openGroups, attrKey)
				currentIndent++
			} else {
				attrValue, _ = strings.CutPrefix(attrValue, " ")
				subEntry[attrKey] = attrValue
			}
		}
	}

	return entry, nil
}

func getSubEntry(entry map[string]any, openGroups []string) map[string]any {
	for _, group := range openGroups {
		var subEntry map[string]any

		candidate, ok := entry[group]
		if ok {
			subEntry, ok = candidate.(map[string]any)
			if !ok {
				return entry
			}
		} else {
			subEntry = make(map[string]any)
			entry[group] = subEntry
		}

		entry = subEntry
	}

	return entry
}
