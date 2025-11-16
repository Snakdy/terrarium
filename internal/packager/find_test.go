package packager

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	t.Run("real file works", func(t *testing.T) {
		ok, err := detectFile(ctx, "", "find.go")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("no file return false without an error", func(t *testing.T) {
		ok, err := detectFile(ctx, "", "this-does-not-exist.txt")
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
