package thoth

import (
	"context"
	"runtime/debug"

	"github.com/expr-thoth/expr-thoth/core"
)

func recoverPanic(ctx context.Context, log core.ThothLogger) {
	if r := recover(); r != nil {
		log.ErrorMetrics(ctx, "Thoth.Panic", "msg: %v\nstack: %v\n", r, string(debug.Stack()))
	}
}
