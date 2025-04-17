package handler

import (
	"Go-Chat/api/rpc"
	"Go-Chat/common/e"
	"Go-Chat/proto"
	"Go-Chat/tools"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

type FormLogin struct {
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
}

func Login(c *gin.Context) {
	var formLogin FormLogin
	if err := c.ShouldBindBodyWith(&formLogin, binding.JSON); err != nil {
		tools.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req := &proto.LoginRequest{
		Name:     formLogin.UserName,
		Password: formLogin.Password,
	}

	// 调用task层的login处理
	code, authToken, msg := rpc.RpcLogicObj.Login(req)
	if code == e.CodeFail || authToken == "" {
		tools.ErrorResponse(c, http.StatusUnauthorized, msg)
		return
	}
	tools.SuccessResponse(c, http.StatusOK, authToken)
}

type FormRegister struct {
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
}

func Register(c *gin.Context) {
	var formRegister FormRegister
	if err := c.ShouldBindBodyWith(&formRegister, binding.JSON); err != nil {
		tools.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req := &proto.RegisterRequest{
		Name:     formRegister.UserName,
		Password: formRegister.Password,
	}
	code, authToken, msg := rpc.RpcLogicObj.Register(req)
	if code == e.CodeFail || authToken == "" {
		tools.ErrorResponse(c, http.StatusUnauthorized, msg)
		return
	}
	tools.SuccessResponse(c, http.StatusOK, authToken)
}

type FormLogout struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

func Logout(c *gin.Context) {
	var formLogout FormLogout
	if err := c.ShouldBindBodyWith(&formLogout, binding.JSON); err != nil {
		tools.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	code := rpc.RpcLogicObj.Logout(&proto.LogoutRequest{AuthToken: formLogout.AuthToken})
	if code == e.CodeFail {
		tools.ErrorResponse(c, http.StatusForbidden, "logout fail!")
		return
	}
	tools.SuccessResponse(c, http.StatusOK, "logout ok!")
}
