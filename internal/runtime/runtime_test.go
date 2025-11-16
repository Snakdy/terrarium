package runtime_test

import (
	"context"
	"testing"

	"github.com/Snakdy/terrarium/internal/runtime"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	t.Run("successful command returns no error", func(t *testing.T) {
		err := runtime.Run(ctx, "", "sh", "true")
		assert.NoError(t, err)
	})
	t.Run("failed command returns error", func(t *testing.T) {
		err := runtime.Run(ctx, "", "sh", "false")
		assert.Error(t, err)
	})
}
