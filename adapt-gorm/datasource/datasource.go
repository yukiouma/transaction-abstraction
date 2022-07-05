package datasource

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/AkiOuma/transaction-abstraction/adapt-gorm/datasource/model"
	"github.com/AkiOuma/transaction-abstraction/adapt-gorm/domain/repository"
	"github.com/AkiOuma/transaction-abstraction/adapt-gorm/domain/valueobject"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Datasource struct {
	data *gorm.DB
}

var _ repository.Repository = (*Datasource)(nil)

func NewDatasource(data *gorm.DB) repository.Repository {
	return &Datasource{
		data: data,
	}
}

func NewGormDB(addr string) *gorm.DB {
	sqlDB, err := sql.Open("mysql", addr)
	if err != nil {
		log.Fatal(err)
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
		},
	)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal(err)
	}
	return gormDB
}

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	return NewDatasource(d.data.Begin()), nil
}

func (d *Datasource) CreateUser(ctx context.Context, name ...string) ([]int, error) {
	creation := make([]model.User, 0, len(name))
	for _, v := range name {
		creation = append(creation, model.User{Name: v})
	}
	if err := d.data.Create(&creation).Error; err != nil {
		return nil, err
	}
	userId := make([]int, 0, len(creation))
	for _, v := range creation {
		userId = append(userId, int(v.ID))
	}
	return userId, nil
}

func (d *Datasource) CreateUserRole(ctx context.Context, data ...*valueobject.UserRole) error {
	creation := make([]model.UserRole, 0, len(data)*2)
	for _, user := range data {
		for _, role := range user.RoleId {
			creation = append(creation, model.UserRole{UserId: uint(user.UserId), RoleId: uint(role)})
		}
	}
	return d.data.Create(&creation).Error
}

func (d *Datasource) EndTx(ctx context.Context, errs ...error) error {
	for _, v := range errs {
		if v != nil {
			d.data.Rollback()
			return v
		}
	}
	d.data.Commit()
	return nil
}
