package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gateway/dao"
	"gateway/dto"
	"gateway/middleware"
	"gateway/public"
	"gateway/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ServiceController struct{}

func ServiceRegister(r *gin.RouterGroup) {
	service := new(ServiceController)
	r.GET("/list", service.ServiceList)
	r.DELETE("/:id", service.ServiceDelete)
	r.GET("/:id", service.ServiceDetail)
	r.GET("/:id/stats", service.ServiceStats)
	r.POST("/http", service.AddHTTPService)
	r.PUT("/http/:id", service.UpdateHTTPService)
	r.POST("/tcp", service.AddTCPService)
	r.PUT("/tcp/:id", service.UpdateTCPService)
	r.POST("/grpc", service.AddGrpcService)
	r.PUT("/grpc/:id", service.UpdateGrpcService)
}

// ServiceList
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page query int true "页数"
// @Param size query int true "每页个数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/list [get]
func (s *ServiceController) ServiceList(ctx *gin.Context) {
	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	service := new(dao.ServiceInfo)
	list, count, err := service.ServiceList(ctx, db, params)
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}

	resList := make([]*dto.ServiceListItemOutput, 0)
	for _, info := range list {
		serviceDetail, err := info.ServiceDetail(ctx, db, info)
		if err != nil {
			middleware.ResponseError(ctx, 10004, err)
			return
		}
		serviceAddr := ""
		clusterIP := utils.GetStringConf("base.cluster.cluster_ip")
		clusterPort := utils.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := utils.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}

		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}

		ipList := serviceDetail.LoadBalance.GetIPListByModel()

		sio := &dto.ServiceListItemOutput{
			ID:          info.ID,
			ServiceName: info.ServiceName,
			ServiceDesc: info.ServiceDesc,
			LoadType:    info.LoadType,
			ServiceAddr: serviceAddr,
			Qpd:         0,
			Qps:         0,
			TotalNode:   len(ipList),
		}

		resList = append(resList, sio)
	}

	data := &dto.ServiceListOutput{
		Total: count,
		List:  resList,
	}
	middleware.ResponseSuccess(ctx, data)
}

// ServiceDelete
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/{id}
// @Accept  json
// @Produce  json
// @Param id path int true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/{id} [delete]
func (s *ServiceController) ServiceDelete(ctx *gin.Context) {

	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	service := &dao.ServiceInfo{}
	err = service.Find(ctx, db, &dao.ServiceInfo{
		ID: id,
	})
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}

	service.IsDelete = 1
	err = service.Save(ctx, db)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}

	middleware.ResponseSuccess(ctx, "success")
}

// AddHTTPService
// @Summary 添加HTTP服务
// @Description 管添加HTTP服务
// @Tags 服务管理
// @ID /service/http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/http [post]
func (*ServiceController) AddHTTPService(ctx *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10002, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if err = serviceInfo.Find(ctx, db, serviceInfo); err == nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10004, errors.New("服务已存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err := httpUrl.Find(ctx, db, httpUrl); err == nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10005, errors.New("服务接入前缀或域名已存在"))
		return
	}

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10006, err)
		return
	}

	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfer: params.HeaderTransfer,
	}
	if err := httpRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}

	lb := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := lb.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10009, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
}

// UpdateServiceHTTP
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/http/{id}
// @Accept  json
// @Produce  json
// @Param id path int64 true "ID"
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/http/{id} [put]
func (*ServiceController) UpdateHTTPService(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10002, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ID: id, ServiceName: params.ServiceName}
	err = serviceInfo.Find(ctx, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10004, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10005, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10006, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfer = params.HeaderTransfer
	if err := httpRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}

	lb := serviceDetail.LoadBalance
	lb.RoundType = params.RoundType
	lb.IpList = params.IpList
	lb.WeightList = params.WeightList
	lb.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	lb.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	lb.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	lb.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := lb.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10009, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
}

// ServiceDetail
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/{id}
// @Accept  json
// @Produce  json
// @Param id path int64 true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/{id} [get]
func (*ServiceController) ServiceDetail(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: id}
	err = serviceInfo.Find(ctx, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}
	middleware.ResponseSuccess(ctx, serviceDetail)
}

// ServiceStats
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/{id}/stats
// @Accept  json
// @Produce  json
// @Param id query int64 true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatsOutput} "success"
// @Router /service/{id}/stats [get]
func (*ServiceController) ServiceStats(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	//读取基本信息
	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ID: id}
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}

	counter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
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

// AddTCPService
// @Summary TCP服务添加
// @Description TCP服务添加
// @Tags 服务管理
// @ID /service/tcp
// @Accept  json
// @Produce  json
// @Param body body dto.AddTCPServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/tcp [post]
func (*ServiceController) AddTCPService(ctx *gin.Context) {
	params := &dto.AddTCPServiceInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if err := infoSearch.Find(ctx, db, infoSearch); err == nil {
		middleware.ResponseError(ctx, 10003, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TCPRule{
		Port: params.Port,
	}
	if err = tcpRuleSearch.Find(ctx, db, tcpRuleSearch); err == nil {
		middleware.ResponseError(ctx, 10004, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(ctx, db, grpcRuleSearch); err == nil {
		middleware.ResponseError(ctx, 10005, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10006, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db = db.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}

	httpRule := &dao.TCPRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	if err := httpRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10009, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10010, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
	return
}

// UpdateTCPService
// @Summary TCP服务更新
// @Description TCP服务更新
// @Tags 服务管理
// @ID /service/tcp/{id}
// @Accept  json
// @Produce  json
// @Param id query int64 true "服务ID"
// @Param body body dto.UpdateTCPServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/tcp/{id} [put]
func (*ServiceController) UpdateTCPService(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	params := &dto.UpdateTCPServiceInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10003, err)
		return
	}

	service := &dao.ServiceInfo{
		ID: id,
	}
	detail, err := service.ServiceDetail(ctx, db, service)
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}

	db = db.Begin()
	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10005, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10006, err)
		return
	}

	tcpRule := &dao.TCPRule{}
	if detail.TCPRule != nil {
		tcpRule = detail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
	return
}

// AddGrpcService
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/grpc
// @Accept  json
// @Produce  json
// @Param body body dto.AddGrpcServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/grpc [post]
func (*ServiceController) AddGrpcService(ctx *gin.Context) {
	params := &dto.AddGrpcServiceInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10001, err)
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if err = infoSearch.Find(ctx, db, infoSearch); err == nil {
		middleware.ResponseError(ctx, 10003, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TCPRule{
		Port: params.Port,
	}
	if err = tcpRuleSearch.Find(ctx, db, tcpRuleSearch); err == nil {
		middleware.ResponseError(ctx, 10004, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(ctx, db, grpcRuleSearch); err == nil {
		middleware.ResponseError(ctx, 10005, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10006, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db = db.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfer: params.HeaderTransfer,
	}
	if err := grpcRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10009, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10010, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
	return
}

// UpdateGrpcService
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/grpc/{id}
// @Accept  json
// @Produce  json
// @Param id query int64 true "服务ID"
// @Param body body dto.UpdateGrpcServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/grpc/{id} [put]
func (*ServiceController) UpdateGrpcService(ctx *gin.Context) {
	ids := ctx.Param("id")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		middleware.ResponseError(ctx, 10001, fmt.Errorf("id invalid"))
		return
	}

	params := &dto.UpdateGrpcServiceInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 10002, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 10003, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db, err := utils.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 10004, err)
		return
	}
	db = db.Begin()
	service := &dao.ServiceInfo{
		ID: id,
	}
	detail, err := service.ServiceDetail(ctx, db, service)
	if err != nil {
		middleware.ResponseError(ctx, 10005, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10006, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10007, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcRule = detail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfer = params.HeaderTransfer
	if err := grpcRule.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10008, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, db); err != nil {
		db.Rollback()
		middleware.ResponseError(ctx, 10009, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(ctx, "success")
	return
}
