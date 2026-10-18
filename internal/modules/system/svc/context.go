package svc

import (
	"github.com/surge-go/aurora/core/logger"
	"gorm.io/gorm"
)

type ServiceContext struct {
	DB  *gorm.DB
	Log *logger.Logger
}

func NewServiceContext(db *gorm.DB, log *logger.Logger) *ServiceContext {
	return &ServiceContext{
		DB:  db,
		Log: log,
	}
}
