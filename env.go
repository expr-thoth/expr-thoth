package thoth

import (
	"context"

	"github.com/expr-thoth/expr-thoth/core"
)

func MakeEnv(ctx context.Context) map[string]any {
	state := &core.ThothState{}
	env := map[string]any{}
	sys := core.NewSystem(ctx, state, env)
	env["Sys"] = sys
	env["ThothState"] = state
	return env
}
