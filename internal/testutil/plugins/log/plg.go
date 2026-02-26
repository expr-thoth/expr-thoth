package log

import (
	"context"

	"github.com/expr-thoth/expr-thoth/core"
)

type PlgLog struct {
	ctx context.Context
}

func Create(ctx context.Context, args ...any) core.Plugin {
	return &PlgLog{}
}

func (p *PlgLog) Init(args ...any) error {
	return nil
}

func (p *PlgLog) Log(msg string) bool {
	core.DefaultSimpleThothLogger.InfoMetrics(p.ctx, "ThothLog", "%s", msg)
	return true
}
