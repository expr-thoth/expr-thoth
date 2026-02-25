package thoth_test

import (
	"context"
	"sync"
	"testing"

	"github.com/expr-thoth/expr-thoth/internal/testutil"
)

func BenchmarkThothBase(b *testing.B) {
	ctx := context.Background()
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}
	// thoth.Execute(req)

	b.ResetTimer()
	for b.Loop() {
		thoth.Execute(ctx, req)
	}
}

func BenchmarkThothC(b *testing.B) {
	var wg sync.WaitGroup
	ctx := context.Background()
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}
	thoth.Execute(ctx, req)

	b.ResetTimer()
	for b.Loop() {
		wg.Add(1)
		// 直接Go实际上性能相当好, 优于手动控制协程+通道发送数据
		go func() {
			defer wg.Done()
			thoth.Execute(ctx, req)
		}()
	}

	wg.Wait()
}
