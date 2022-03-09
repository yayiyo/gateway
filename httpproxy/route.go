package httpproxy

import (
	"gateway/controller"
	"gateway/httpproxy/middleware"
	. "gateway/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	oauth := router.Group("/oauth")
	oauth.Use(TranslationMiddleware())
	controller.OAuthRegister(oauth)

	router.Use(
		middleware.HTTPAccessModeMiddleware(),
		middleware.HTTPFlowCountMiddleware(),
		middleware.HTTPFlowLimitMiddleware(),
		middleware.HTTPJwtAuthTokenMiddleware(),
		middleware.HTTPJwtFlowCountMiddleware(),
		middleware.HTTPJwtFlowLimitMiddleware(),
		middleware.HTTPWhiteListMiddleware(),
		middleware.HTTPBlackListMiddleware(),
		middleware.HTTPHeaderTransferMiddleware(),
		middleware.HTTPStripUriMiddleware(),
		middleware.HTTPUrlRewriteMiddleware(),
		middleware.HTTPReverseProxyMiddleware())

	return router
}
