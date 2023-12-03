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

func TestJSON(t *testing.T) {
	user := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Name:  "hermannm",
		Email: "test@example.com",
	}

	output := getLogOutput(nil, func() {
		log.Info("user created", log.JSON("user", user))
	})

	assertContains(
		t,
		output,
		`"user created"`,
		`"user":{"name":"hermannm","email":"test@example.com"}`,
	)
}
