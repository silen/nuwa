package tools

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/silen/nuwa/pkg/logs"
)

func Defer(ctx context.Context) {
	r := recover()
	if r == nil {
		return
	}
	err := fmt.Sprintf("%v", r)
	stack := string(debug.Stack())
	logs.WithContext(ctx).Error("err:" + err + "debug_stack:" + stack)
}
