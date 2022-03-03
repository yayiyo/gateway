package dto

import (
	"time"

	"gateway/public"
	"github.com/gin-gonic/gin"
)

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}

type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"账号" example:"admin" validate:"required,valid_username"`
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

func (param *AdminLoginInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, param)
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"token" example:"123456" validate:""`
}

type AdminInfoOutput struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type ChangePwdInput struct {
	OldPwd   string `json:"old_pwd" form:"old_pwd" comment:"旧密码" example:"123456" validate:"required"`
	Password string `json:"password" form:"password" comment:"新密码" example:"abcdef" validate:"required"`
}

func (param *ChangePwdInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, param)
}
