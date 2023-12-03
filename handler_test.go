package devlog_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
)

// Tests our handler against the standard library test suite for structured logging handlers.
func TestSlog(t *testing.T) {
	var buf bytes.Buffer

	err := slogtest.TestHandler(
		devlog.NewHandler(&buf, &devlog.Options{DisableColors: true}),
		func() []map[string]any {
			entries, err := parseLogEntries(buf.String())
			if err != nil {
				t.Fatal(err)
			}
			return entries
		},
	)

	if err != nil {
		t.Error(err)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(devlog.NewHandler(&buf, &devlog.Options{DisableColors: true}))

	user := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Name:  "hermannm",
		Email: "test@example.com",
	}

	logger.Info("user created", log.JSON("user", user))

	expectedOutput := `user: {
    "name": "hermannm",
    "email": "test@example.com"
  }`

	assertContains(t, buf.String(), expectedOutput)
}

func TestListAttributes(t *testing.T) {
	type testStruct struct {
		text string
	}

	testCases := []struct {
		attribute      slog.Attr
		expectedOutput string
	}{
		{
			attribute: slog.Any("stringList", []string{"test1", "test2", "test3"}),
			expectedOutput: `  stringList:
    - test1
    - test2
    - test3`,
		},
		{
			attribute: slog.Any("structList", []testStruct{{"test1"}, {"test2"}}),
			expectedOutput: `  structList:
    - {test1}
    - {test2}`,
		},
		{
			attribute: slog.Any("multilineStringList", []string{`multiline
string 1`, `multiline
string 2`}),
			expectedOutput: `  multilineStringList:
    - multiline
      string 1
    - multiline
      string 2`,
		},
		{
			attribute: slog.Any("multilineStructList", []testStruct{{`multiline
string 1`}, {`multiline
string 2`}}),
			expectedOutput: `  multilineStructList:
    - {multiline
      string 1}
    - {multiline
      string 2}`,
		},
		{
			attribute:      slog.Any("singleListItem", []string{"single"}),
			expectedOutput: "  singleListItem: single",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.attribute.Key, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(devlog.NewHandler(&buf, &devlog.Options{DisableColors: true}))
			logger.Info("", testCase.attribute)

			output := buf.String()
			t.Log(output)
			assertContains(t, output, testCase.expectedOutput)
		})
	}
}

func assertContains(t *testing.T, output string, expectedOutput string) {
	t.Helper()

	if !strings.Contains(output, expectedOutput) {
		t.Errorf(`log did not contain expected output
got:
----------------------------------------
%s
----------------------------------------

want:
----------------------------------------
%s
----------------------------------------
`, output, expectedOutput)
	}
}

// slogtest.TestHandler requires us to parse our logging output to a []map[string]any.
func parseLogEntries(data string) ([]map[string]any, error) {
	entryStrings := strings.Split(data, "\n[")
	entryStrings[0], _ = strings.CutPrefix(entryStrings[0], "[")
	last := len(entryStrings) - 1
	entryStrings[last], _ = strings.CutSuffix(entryStrings[last], "\n")

	entries := make([]map[string]any, 0, len(entryStrings))
	for _, entryString := range entryStrings {
		if err := parseLogEntry(&entries, entryString, true); err != nil {
			return nil, fmt.Errorf("failed to parse log entry number %d: %w", len(entries)+1, err)
		}
	}

	return entries, nil
}

func parseLogEntry(entries *[]map[string]any, entryString string, includeTime bool) error {
	entry := make(map[string]any)

	if includeTime {
		split := strings.SplitN(entryString, "] ", 2)
		timeString := split[0]
		entryString = split[1]

		time, err := time.Parse(time.DateTime, timeString)
		if err != nil {
			return fmt.Errorf("failed to parse time: %w", err)
		}
		entry[slog.TimeKey] = time
	}

	split := strings.SplitN(entryString, ": ", 2)
	levelString := split[0]
	entryString = split[1]

	entry[slog.LevelKey] = levelString

	split = strings.SplitN(entryString, "\n", 2)
	msg := split[0]
	hasAttributes := len(split) == 2

	entry[slog.MessageKey] = msg

	var entriesWithoutTime []string
	if hasAttributes {
		entryString = split[1]
		for {
			if i := strings.Index(entryString, "\n"+slog.LevelInfo.String()); i >= 0 {
				currentEntry, nextEntry := entryString[:i], entryString[i+1:]
				entryString = currentEntry
				entriesWithoutTime = append(entriesWithoutTime, nextEntry)
			} else {
				break
			}
		}

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

	*entries = append(*entries, entry)

	for _, entryString := range entriesWithoutTime {
		if err := parseLogEntry(entries, entryString, false); err != nil {
			return err
		}
	}

	return nil
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
