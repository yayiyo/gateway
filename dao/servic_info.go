package dao

import (
	"time"

	"gateway/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (*ServiceInfo) TableName() string {
	return "service_info"
}

func (t *ServiceInfo) ServiceList(ctx *gin.Context, db *gorm.DB, input *dto.ServiceListInput) ([]*ServiceInfo, int64, error) {
	var (
		list  = make([]*ServiceInfo, 0)
		total int64
		err   error
	)

	offset := (input.Page - 1) * input.Size

	query := db.WithContext(ctx).Table(t.TableName()).Where("is_delete=0")

	if input.Info != "" {
		query.Where(`(service_name like %?%) or (service_desc like %?%)`, input.Info, input.Info)
	}

	if err = query.Limit(input.Size).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	if err = query.Limit(input.Size).Offset(offset).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return list, int64(len(list)), nil
}

func (t *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	err := db.WithContext(c).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (t *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	return db.WithContext(c).Save(t).Error
}
