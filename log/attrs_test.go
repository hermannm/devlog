package log_test

import (
	"testing"

	"hermannm.dev/devlog/log"
)

func TestString(t *testing.T) {
	type username string
	name := username("hermannm")

	output := getLogOutput(nil, func() {
		log.Info("user created", log.String("name", name))
	})

	assertContains(t, output, `"user created"`, `"name":"hermannm"`)
}
