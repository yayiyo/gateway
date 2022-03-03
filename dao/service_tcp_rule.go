package dao

import (
	"gateway/public"
	"github.com/echaser/gorm"
	"github.com/gin-gonic/gin"
)

type TCPRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (t *TCPRule) TableName() string {
	return "tcp_rule"
}

func (t *TCPRule) Find(c *gin.Context, tx *gorm.DB, search *TCPRule) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(t).Error
}

func (t *TCPRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *TCPRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]TCPRule, int64, error) {
	var list []TCPRule
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
