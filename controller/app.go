package controller

import (
	"fmt"
	"strconv"
	"time"

	"gateway/dao"
	"gateway/dto"
	"gateway/middleware"
	"gateway/public"
	"gateway/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type APPController struct{}

//APPControllerRegister 路由注册
func APPRegister(router *gin.RouterGroup) {
	admin := APPController{}
	router.GET("", admin.APPList)
	router.GET("/:id", admin.APPDetail)
	router.GET("/:id/stats", admin.AppStatistics)
	router.DELETE("/:id", admin.APPDelete)
	router.POST("", admin.AppAdd)
	router.PUT("/:id", admin.AppUpdate)
}

// APPList
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /apps
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param size query string true "每页多少条"
// @Param page query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /apps [get]
func (*APPController) APPList(ctx *gin.Context) {
	params := &dto.APPListInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	info := &dao.App{}
	list, total, err := info.APPList(ctx, utils.GORMDefaultPool, params)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	outputList := make([]*dto.APPListItemOutput, 0)
	for _, item := range list {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(ctx, 10003, err)
			ctx.Abort()
			return
		}
		outputList = append(outputList, &dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.APPListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(ctx, output)
	return
}

// APPDetail
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /apps/{id}
// @Accept  json
// @Produce  json
// @Param id query int true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /apps/{id} [get]
func (*APPController) APPDetail(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}
	app := &dao.App{
		ID: id,
	}
	err = app.Find(ctx, utils.GORMDefaultPool, app)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	middleware.ResponseSuccess(ctx, app)
	return
}

// APPDelete
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /apps/{id}
// @Accept  json
// @Produce  json
// @Param id query int true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps/{id} [DELETE]
func (*APPController) APPDelete(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}
	app := &dao.App{
		ID: id,
	}
	err = app.Find(ctx, utils.GORMDefaultPool, app)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	app.IsDelete = 1
	if err := app.Save(ctx, utils.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "success")
	return
}

// AppAdd
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /apps
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps [post]
func (*APPController) AppAdd(ctx *gin.Context) {
	params := &dto.APPAddHttpInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if err := search.Find(ctx, utils.GORMDefaultPool, search); err == nil {
		middleware.ResponseError(ctx, 10002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	tx := utils.GORMDefaultPool
	info := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "success")
	return
}

// AppUpdate
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /apps/{id}
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps/{id} [put]
func (*APPController) AppUpdate(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	params := &dto.APPUpdateHttpInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	app := &dao.App{
		ID: id,
	}
	err = app.Find(ctx, utils.GORMDefaultPool, app)
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	app.Name = params.Name
	app.Secret = params.Secret
	app.WhiteIPS = params.WhiteIPS
	app.Qps = params.Qps
	app.Qpd = params.Qpd
	if err := app.Save(ctx, utils.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}
	middleware.ResponseSuccess(ctx, "success")
	return
}

// AppStatistics
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /apps/{id}/stats
// @Accept  json
// @Produce  json
// @Param id query int true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /apps/{id}/stats [get]
func (*APPController) AppStatistics(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	app := &dao.App{
		ID: id,
	}

	err = app.Find(ctx, utils.GORMDefaultPool, app)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := make([]int64, 0)
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + app.AppID)
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		ctx.Abort()
		return
	}
	currentTime := time.Now()
	for i := 0; i <= time.Now().In(utils.TimeLocation).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, utils.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStats := make([]int64, 0)
	yesterdayTime := currentTime.Add(-1 * time.Hour * 24)
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, utils.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayStats = append(yesterdayStats, hourData)
	}
	stats := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStats,
	}
	middleware.ResponseSuccess(ctx, stats)
	return
}
