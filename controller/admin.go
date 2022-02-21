package controller

import (
	"encoding/json"
	"errors"
	"time"

	"gateway/dao"
	"gateway/dto"
	"gateway/middleware"
	"gateway/public"
	"gateway/utils"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminLoginController struct{}

func AdminLoginRegister(r *gin.RouterGroup) {
	adminLogin := new(AdminLoginController)
	r.POST("/login", adminLogin.AdminLogin)
	r.POST("/logout", adminLogin.AdminLogout)
}

// AdminLogin
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLogin *AdminLoginController) AdminLogin(ctx *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(ctx, db, params)
	if err != nil {
		middleware.ResponseError(ctx, 1003, err)
		return
	}

	sessionInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}

	sb, err := json.Marshal(sessionInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}

	sess := sessions.Default(ctx)
	sess.Set(public.AdminSessionInfoKey, string(sb))
	sess.Save()

	data := &dto.AdminLoginOutput{
		Token: "123456",
	}
	middleware.ResponseSuccess(ctx, data)
}

// AdminLogout
// @Summary 管理员登出
// @Description 管理员登出
// @Tags 管理员
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [post]
func (adminLogin *AdminLoginController) AdminLogout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Delete(public.AdminSessionInfoKey)
	sess.Save()
	middleware.ResponseSuccess(ctx, "success")
}

type AdminController struct{}

func AdminRegister(r *gin.RouterGroup) {
	admin := new(AdminController)
	r.GET("/info", admin.AdminInfo)
	r.POST("/change_pwd", admin.ChangePwd)
}

// AdminInfo
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员
// @ID /admin/info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/info [get]
func (a *AdminController) AdminInfo(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sbi := sess.Get(public.AdminSessionInfoKey)
	sessionInfo := &dto.AdminSessionInfo{}
	err := json.Unmarshal([]byte(sbi.(string)), sessionInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	data := &dto.AdminInfoOutput{
		ID:           sessionInfo.ID,
		Name:         sessionInfo.UserName,
		LoginTime:    sessionInfo.LoginTime,
		Avatar:       "",
		Introduction: "I am the admin manager.",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(ctx, data)
}

// ChangePwd
// @Summary 修改密码
// @Description 修改密码
// @Tags 管理员
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (a *AdminController) ChangePwd(ctx *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	sess := sessions.Default(ctx)
	sbi := sess.Get(public.AdminSessionInfoKey)
	sessionInfo := &dto.AdminSessionInfo{}
	err := json.Unmarshal([]byte(sbi.(string)), sessionInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.Find(ctx, db, &dao.Admin{
		UserName: sessionInfo.UserName,
	})
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}

	oldPwd := public.GenSaltPassword(params.OldPwd, admin.Salt)
	if oldPwd != admin.Password {
		middleware.ResponseError(ctx, 10005, errors.New("旧密码不正确"))
		return
	}

	newPwd := public.GenSaltPassword(params.Password, admin.Salt)
	admin.Password = newPwd
	err = admin.Save(ctx, db)
	if err != nil {
		middleware.ResponseError(ctx, 10006, err)
		return
	}

	middleware.ResponseSuccess(ctx, "success")
}
