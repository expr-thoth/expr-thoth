package core

import (
	"context"
	"fmt"
)

type ThothLogger interface {
	ErrorMetrics(ctx context.Context, metricsName string, logContent string, args ...any)
	WarnMetrics(ctx context.Context, metricsName string, logContent string, args ...any)
	InfoMetrics(ctx context.Context, metricsName string, logContent string, args ...any)
	DebugMetrics(ctx context.Context, metricsName string, logContent string, args ...any)
}

// 日志默认简单实现 输出到控制台
type SimpleThothLogger struct{}

func (SimpleThothLogger) ErrorMetrics(ctx context.Context, metricsName string, logContent string, args ...any) {
	fmt.Printf("[ERROR] [%s] %s\n", metricsName, fmt.Sprintf(logContent, args...))
}

func (SimpleThothLogger) WarnMetrics(ctx context.Context, metricsName string, logContent string, args ...any) {
	fmt.Printf("[Warn] [%s] %s\n", metricsName, fmt.Sprintf(logContent, args...))
}

func (SimpleThothLogger) InfoMetrics(ctx context.Context, metricsName string, logContent string, args ...any) {
	fmt.Printf("[Info] [%s] %s\n", metricsName, fmt.Sprintf(logContent, args...))
}

func (SimpleThothLogger) DebugMetrics(ctx context.Context, metricsName string, logContent string, args ...any) {
	fmt.Printf("[Debug] [%s] %s\n", metricsName, fmt.Sprintf(logContent, args...))
}
