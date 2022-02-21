package dto

import (
	"gateway/public"
	"github.com/gin-gonic/gin"
)

type ServiceListInput struct {
	Info string `json:"info" form:"info" comment:"关键词" example:"" validate:""`
	Page int    `json:"page" form:"page" comment:"页数" example:"1" validate:""`
	Size int    `json:"size" form:"size" comment:"每页个数" example:"20" validate:""`
}

func (param *ServiceListInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, param)
}

type ServiceListItemOutput struct {
	ID          int64  `json:"id" form:"id"`                     //id
	ServiceName string `json:"service_name" form:"service_name"` //服务名称
	ServiceDesc string `json:"service_desc" form:"service_desc"` //服务描述
	LoadType    int    `json:"load_type" form:"load_type"`       //类型
	ServiceAddr string `json:"service_addr" form:"service_addr"` //服务地址
	Qps         int64  `json:"qps" form:"qps"`                   //qps
	Qpd         int64  `json:"qpd" form:"qpd"`                   //qpd
	TotalNode   int    `json:"total_node" form:"total_node"`     //节点数
}

type ServiceListOutput struct {
	Total int64                    `json:"total" form:"total" comment:"总数" example:"" validate:""` //总数
	List  []*ServiceListItemOutput `json:"list" form:"list" comment:"列表" example:"" validate:""`   //列表
}
