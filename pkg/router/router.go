package router

import (
	"github.com/gin-gonic/gin"
	"oauth2/pkg/controller"
)

func Setup(r *gin.Engine) {
	r.GET("/authorize", controller.AuthorizeHandler)
	r.POST("/login", controller.LoginHandler)
	r.POST("/logout", controller.LogoutHandler)
	r.POST("/token", controller.TokenHandler)
	r.POST("/verify", controller.VerifyHandler)
	// 静态文件服务，使用项目根目录下的 static 目录
	r.Static("/static", controller.GetTemplatePath("static"))
	r.GET("/login", controller.GETloginHandler)
	r.GET("/", controller.NotFoundHandler)
}
