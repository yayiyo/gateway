package dao

import (
	"time"

	"gateway/dto"
	"gateway/public"
	"github.com/echaser/gorm"
	"github.com/gin-gonic/gin"
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

func (s *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	if search.ServiceName == "" {
		err := s.Find(c, tx, search)
		if err != nil {
			return nil, err
		}
		search = s
	}
	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err := httpRule.Find(c, tx, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TCPRule{ServiceID: search.ID}
	err = tcpRule.Find(c, tx, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: search.ID}
	err = grpcRule.Find(c, tx, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: search.ID}
	accessControl, err = accessControl.Find(c, tx, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: search.ID}
	loadBalance, err = loadBalance.Find(c, tx, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	detail := &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail, nil
}

func (s *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]*dto.DashServiceStatsItemOutput, error) {
	list := make([]*dto.DashServiceStatsItemOutput, 0)
	query := tx.SetCtx(public.GetGinTraceContext(c))
	if err := query.Table(s.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServiceInfo) ServiceList(ctx *gin.Context, db *gorm.DB, input *dto.ServiceListInput) ([]*ServiceInfo, int64, error) {
	var (
		list  = make([]*ServiceInfo, 0)
		total int64
		err   error
	)

	offset := (input.Page - 1) * input.Size

	query := db.SetCtx(public.GetGinTraceContext(ctx)).Table(s.TableName())

	if input.Info != "" {
		query = query.Where("(service_name like ? or service_desc like ?) and is_delete=0", "%"+input.Info+"%", "%"+input.Info+"%")
	} else {
		query = query.Where("is_delete=0")
	}

	if err = query.Limit(input.Size).Offset(offset).Find(&list).Order("id desc").Error; err != nil {
		return nil, 0, err
	}

	if err = query.Limit(input.Size).Offset(offset).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) error {
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(s).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(c)).Save(s).Error
}

func (s *ServiceInfo) Delete(c *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(c)).Delete(s).Error
}
