package panic

import (
	"context"

	"github.com/expr-thoth/expr-thoth/core"
)

type PlgPanic struct {
}

func Create(ctx context.Context, args ...any) core.Plugin {
	return PlgPanic{}
}

func (PlgPanic) Init(args ...any) error {
	return nil
}

func (PlgPanic) Panic(msg string) bool {
	panic(msg)
}
