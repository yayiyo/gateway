package public

import (
	"context"

	"gateway/utils"
	"github.com/gin-gonic/gin"
)

//错误日志
func ContextWarning(c context.Context, dltag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*utils.TraceContext)
	if !ok {
		traceContext = utils.NewTrace()
	}
	utils.Log.TagWarn(traceContext, dltag, m)
}

//错误日志
func ContextError(c context.Context, dltag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*utils.TraceContext)
	if !ok {
		traceContext = utils.NewTrace()
	}
	utils.Log.TagError(traceContext, dltag, m)
}

//普通日志
func ContextNotice(c context.Context, dltag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*utils.TraceContext)
	if !ok {
		traceContext = utils.NewTrace()
	}
	utils.Log.TagInfo(traceContext, dltag, m)
}

//错误日志
func ComLogWarning(c *gin.Context, dltag string, m map[string]interface{}) {
	traceContext := GetGinTraceContext(c)
	utils.Log.TagError(traceContext, dltag, m)
}

//普通日志
func ComLogNotice(c *gin.Context, dltag string, m map[string]interface{}) {
	traceContext := GetGinTraceContext(c)
	utils.Log.TagInfo(traceContext, dltag, m)
}

// 从gin的Context中获取数据
func GetGinTraceContext(c *gin.Context) *utils.TraceContext {
	// 防御
	if c == nil {
		return utils.NewTrace()
	}
	traceContext, exists := c.Get("trace")
	if exists {
		if tc, ok := traceContext.(*utils.TraceContext); ok {
			return tc
		}
	}
	return utils.NewTrace()
}

// 从Context中获取数据
func GetTraceContext(c context.Context) *utils.TraceContext {
	if c == nil {
		return utils.NewTrace()
	}
	traceContext := c.Value("trace")
	if tc, ok := traceContext.(*utils.TraceContext); ok {
		return tc
	}
	return utils.NewTrace()
}
