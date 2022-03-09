package controller

import (
	"time"

	"gateway/dao"
	"gateway/dto"
	"gateway/middleware"
	"gateway/public"
	"gateway/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type DashboardController struct{}

func DashboardRegister(group *gin.RouterGroup) {
	service := &DashboardController{}
	group.GET("/panel_group_data", service.PanelGroupData)
	group.GET("/flow_stats", service.FlowStats)
	group.GET("/service_stats", service.ServiceStats)
}

// PanelGroupData
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (*DashboardController) PanelGroupData(ctx *gin.Context) {
	tx, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	_, serviceNum, err := serviceInfo.ServiceList(ctx, tx, &dto.ServiceListInput{Size: 1, Page: 1})
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	app := &dao.App{}
	_, appNum, err := app.APPList(ctx, tx, &dto.APPListInput{Page: 1, Size: 1})
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}
	out := &dto.PanelGroupDataOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		TodayRequestNum: counter.TotalCount,
		CurrentQPS:      counter.QPS,
	}
	middleware.ResponseSuccess(ctx, out)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stats
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatsOutput} "success"
// @Router /dashboard/service_stats [get]
func (*DashboardController) ServiceStats(ctx *gin.Context) {
	tx, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(ctx, tx)
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	legend := make([]string, 0)
	for index, item := range list {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(ctx, 10003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}
	out := &dto.DashServiceStatsOutput{
		Legend: legend,
		Data:   list,
	}
	middleware.ResponseSuccess(ctx, out)
}

// FlowStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/flow_stats
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatsOutput} "success"
// @Router /dashboard/flow_stats [get]
func (*DashboardController) FlowStats(ctx *gin.Context) {
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}
	todayList := make([]int64, 0)
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, utils.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	yesterdayList := make([]int64, 0)
	yesterdayTime := currentTime.Add(-1 * time.Hour * 24)
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, utils.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}
	middleware.ResponseSuccess(ctx, &dto.ServiceStatsOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}
