package devlog

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"hermannm.dev/devlog/errlog"
)

func TestNeedsGroups(t *testing.T) {
	assert.False(t, needsGroups(nil))

	assert.False(t, needsGroups(&Options{ReplaceAttr: nil}))

	assert.False(t, needsGroups(&Options{ReplaceAttr: errlog.ReplaceErrorAttr}))

	assert.True(
		t,
		needsGroups(
			&Options{
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					return attr
				},
			},
		),
	)
}
