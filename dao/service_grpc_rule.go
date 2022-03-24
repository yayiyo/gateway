package dao

import (
	"gateway/public"
	"github.com/echaser/gorm"
	"github.com/gin-gonic/gin"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port           int    `json:"port" gorm:"column:port" description:"端口	"`
	HeaderTransfer string `json:"header_transfer" gorm:"column:header_transfer" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

func (t *GrpcRule) TableName() string {
	return "grpc_rule"
}

func (t *GrpcRule) Find(c *gin.Context, tx *gorm.DB, search *GrpcRule) error {
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(t).Error
	return err
}

func (t *GrpcRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *GrpcRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]GrpcRule, int64, error) {
	var list []GrpcRule
	var count int64
	query := tx.SetCtx(public.GetGinTraceContext(c))
	query = query.Table(t.TableName()).Select("*")
	query = query.Where("service_id=?", serviceID)
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}
