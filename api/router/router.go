package router

import (
	"Go-Chat/api/handler"
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Register() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware()) //跨域
	initUserRouter(r)       // 用户鉴权
	r.NoRoute(func(c *gin.Context) {
		tools.ErrorResponse(c, http.StatusNotFound, "please check request url !")
	})
	return r
}

func initUserRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)
	userGroup.Use(VerifiyJwt())
	{
		userGroup.POST("/logout", handler.Logout)
	}

}

func VerifiyJwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get(config.Conf.Jwt.Name) // `token`

		// 解析获取用户载荷信息
		payLoad, err := tools.ParseToken(token, config.Conf.Jwt.Secret)
		if err != nil {
			code := e.FailReplyCode
			c.JSON(http.StatusUnauthorized, tools.ResponseData{
				Code:    code,
				Message: "Jwt校验失败",
			})
			c.Abort()
			return
		}
		// 在上下文设置载荷信息
		c.Set(e.CurrentId, payLoad.UserId)
		c.Set(e.CurrentName, payLoad.GrantScope)
		c.Next()
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}
