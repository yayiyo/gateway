package dao

import (
	"time"

	"gateway/dto"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"账号"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (a *Admin) TableName() string {
	return "admin"
}

func (a *Admin) LoginCheck(ctx *gin.Context, tx *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {
	adminInfo, err := a.Find(ctx, tx, &Admin{UserName: param.UserName, IsDelete: 0})
	if err != nil {
		return nil, errors.New("用户信息不存在")
	}

	saltPwd := public.GenSaltPassword(param.Password, adminInfo.Salt)
	if saltPwd != adminInfo.Password {
		return nil, errors.New("密码不正确")
	}

	return adminInfo, nil
}

func (a *Admin) Find(c *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	admin := &Admin{}
	err := tx.WithContext(c).Where(search).Find(admin).Error
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (a *Admin) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(a).Error
}
