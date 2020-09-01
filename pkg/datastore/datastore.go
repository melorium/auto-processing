package datastore

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Base struct {
	ID    uint   `json:"id" gorm:"AUTO_INCREMENT"`
	CTime int64  `json:"cTime"`
	MTime int64  `json:"mTime"`
	DTime *int64 `sql:"index" json:"dTime"`
}

func (b *Base) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CTime", time.Now().Unix())
	return nil
}

func (b *Base) BeforeSave(scope *gorm.Scope) (err error) {
	scope.SetColumn("MTime", time.Now().Unix())
	return nil
}

func (b *Base) BeforeDelete(scope *gorm.Scope) (err error) {
	scope.SetColumn("DTime", time.Now().Unix())
	return nil
}
