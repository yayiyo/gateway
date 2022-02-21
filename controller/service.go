package controller

import (
	"gateway/dao"
	"gateway/dto"
	"gateway/middleware"
	"gateway/utils"
	"github.com/gin-gonic/gin"
)

type ServiceController struct{}

func ServiceRegister(r *gin.RouterGroup) {
	service := new(ServiceController)
	r.GET("/list", service.ServiceList)
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
		sio := &dto.ServiceListItemOutput{
			ID:          info.ID,
			ServiceName: info.ServiceName,
			ServiceDesc: info.ServiceDesc,
			LoadType:    info.LoadType,
		}

		resList = append(resList, sio)
	}

	data := &dto.ServiceListOutput{
		Total: count,
		List:  resList,
	}
	middleware.ResponseSuccess(ctx, data)
}
