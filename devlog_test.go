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

	"github.com/stretchr/testify/assert"

	"hermannm.dev/devlog"
)

// Tests our handler against the standard library test suite for structured log handlers.
//
//nolint:thelper // We don't need to mark these as helper functions
func TestSlog(t *testing.T) {
	var buffer bytes.Buffer

	slogtest.Run(
		t,
		func(t *testing.T) slog.Handler {
			buffer.Reset()
			return devlog.NewHandler(
				&buffer, &devlog.Options{
					DisableColors: true,
					TimeFormat:    devlog.TimeFormatFull,
				},
			)
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
		handler := devlog.NewHandler(
			&buffer,
			&devlog.Options{DisableColors: true, TimeFormat: testCase.format},
		)

		record := slog.NewRecord(timeValue, slog.LevelInfo, "Message", 0)
		if err := handler.Handle(context.Background(), record); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		assert.Contains(t, buffer.String(), testCase.expectedOutput)
	}
}

func TestTimeFormatNone(t *testing.T) {
	timeValue, err := time.Parse(time.DateTime, "2024-09-29 10:57:30")
	if err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	handler := devlog.NewHandler(
		&buffer,
		&devlog.Options{DisableColors: true, TimeFormat: devlog.TimeFormatFull},
	)

	if err := handler.Handle(
		context.Background(),
		slog.NewRecord(timeValue, slog.LevelInfo, "Message", 0),
	); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buffer.String()

	assert.NotContains(t, output, "[", "Should not contain brackets")
	assert.NotContains(t, output, "]", "Should not contain brackets")

	assert.NotContains(t, output, "2024-09-29", "Should not contain date")
	assert.NotContains(t, output, "10:57:30", "Should not contain time")
}

type event struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

func TestStructAttr(t *testing.T) {
	event := event{
		ID:   1000,
		Type: "ORDER_UPDATED",
	}

	output := getLogOutput(
		func() {
			slog.Error("Failed to process event", "event", event)
		},
	)

	expectedOutput := `event: {
    "id": 1000,
    "type": "ORDER_UPDATED"
  }`

	assert.Contains(t, output, expectedOutput)
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
		t.Run(
			testCase.attr.Key, func(t *testing.T) {
				output := getLogOutput(
					func() {
						slog.Info("", testCase.attr) //nolint:loggercheck // False positive
					},
				)
				assert.Contains(t, output, testCase.expectedOutput)
			},
		)
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

	output := getLogOutput(
		func() {
			slog.Error("", "cause", errorLog)
		},
	)

	assert.Contains(
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
	output := getLogOutputWithOptions(
		&devlog.Options{AddSource: true},
		func() {
			slog.Info("Test")
		},
	)

	assert.Contains(t, output, "\n  source: hermannm.dev/devlog_test.TestSource")
	assert.Contains(t, output, "devlog_test.go:221")
}

func TestReplaceAttrWithGroups(t *testing.T) {
	type GroupsAndAttr struct {
		groups []string
		attr   slog.Attr
	}

	var replaceCalls []GroupsAndAttr

	_ = getLogOutputWithOptions(
		&devlog.Options{
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				replaceCalls = append(replaceCalls, GroupsAndAttr{groups, attr})
				return attr
			},
		},
		func() {
			slog.Info(
				"Test",
				slog.String("key1", "value1"),
				slog.Group(
					"group1",
					slog.String("key2", "value2"),
					slog.Group(
						"group2",
						slog.String("key3", "value3"),
					),
					slog.String("key4", "value4"),
				),
			)
		},
	)

	assert.Equal(
		t,
		[]GroupsAndAttr{
			{groups: nil, attr: slog.String("key1", "value1")},
			{groups: []string{"group1"}, attr: slog.String("key2", "value2")},
			{groups: []string{"group1", "group2"}, attr: slog.String("key3", "value3")},
			{groups: []string{"group1"}, attr: slog.String("key4", "value4")},
		},
		replaceCalls,
	)
}

func getLogOutput(logFunc func()) string {
	options := &devlog.Options{Level: slog.LevelDebug}
	return getLogOutputWithOptions(options, logFunc)
}

func getLogOutputWithOptions(handlerOptions *devlog.Options, logFunc func()) string {
	// Must disable color output to parse reliably
	handlerOptions.DisableColors = true

	var buffer bytes.Buffer
	slog.SetDefault(slog.New(devlog.NewHandler(&buffer, handlerOptions)))

	logFunc()

	output := buffer.String()

	return output
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
