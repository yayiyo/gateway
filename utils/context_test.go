package utils

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetGinTraceContext(t *testing.T) {
	trace := &TraceContext{
		Trace: Trace{
			TraceId: "sssss",
		},
		CSpanId: "test trace",
	}

	ginCtx := &gin.Context{}
	_ = SetGinTraceContext(ginCtx, trace)
	trace = GetTraceContext(ginCtx)
	t.Log(*trace)

	ctx := context.Background()
	ctx = SetTraceContext(ctx, trace)
	trace = GetTraceContext(ctx)

	t.Log(*trace)
}
