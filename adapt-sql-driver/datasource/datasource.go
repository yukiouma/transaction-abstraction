package datasource

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/AkiOuma/transaction-abstraction/adapt-sql-driver/domain/repository"
	"github.com/AkiOuma/transaction-abstraction/adapt-sql-driver/domain/valueobject"
	_ "github.com/go-sql-driver/mysql"
)

type Datasource struct {
	data *sql.DB
	tx   *sql.Tx
}

var _ repository.Repository = (*Datasource)(nil)

func NewSqlDB(addr string) *sql.DB {
	db, err := sql.Open("mysql", addr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func NewDatasource(data *sql.DB) *Datasource {
	return &Datasource{data: data}
}

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	tx, err := d.data.Begin()
	if err != nil {
		return nil, err
	}
	ds := NewDatasource(d.data)
	ds.tx = tx
	return ds, nil
}

func (d *Datasource) CreateUser(ctx context.Context, name ...string) ([]int, error) {

	// build syntax

	param := make([]interface{}, 0, len(name))
	if len(name) == 0 {
		return nil, nil
	}
	sqlSyntax := "INSERT INTO `users` (`name`) VALUES "
	var sb strings.Builder
	sb.Grow(len(sqlSyntax) + 10*len(name))
	sb.WriteString(sqlSyntax)
	for _, v := range name {
		sb.WriteString("(?),")
		param = append(param, v)
	}
	sqlSyntax = sb.String()
	sqlSyntax = sqlSyntax[:len(sqlSyntax)-1]
	sqlSyntax += ";"
	log.Printf("exec: %s args=%v", sqlSyntax, param)

	// execute creation
	var (
		result sql.Result
		err    error
	)
	if d.tx != nil {
		result, err = d.tx.ExecContext(ctx, sqlSyntax, param...)
	} else {
		result, err = d.data.ExecContext(ctx, sqlSyntax, param...)
	}

	if err != nil {
		return nil, err
	}
	userId := make([]int, 0, len(name))
	total, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	for total > 0 {
		userId = append(userId, int(lastId))
		lastId--
		total--
	}
	return userId, nil
}

func (d *Datasource) CreateUserRole(ctx context.Context, data ...*valueobject.UserRole) error {

	// build syntax

	param := make([]interface{}, 0, len(data)*2)
	if len(data) == 0 {
		return nil
	}
	sqlSyntax := "INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES "
	var sb strings.Builder
	sb.Grow(len(sqlSyntax) + 10*2*len(data))
	sb.WriteString(sqlSyntax)
	for _, user := range data {
		for _, role := range user.RoleId {
			sb.WriteString("(?, ?),")
			param = append(param, user.UserId)
			param = append(param, role)
		}
	}
	sqlSyntax = sb.String()
	sqlSyntax = sqlSyntax[:len(sqlSyntax)-1]
	sqlSyntax += ";"
	log.Printf("exec: %s args=%v", sqlSyntax, param)

	// execute creation

	var (
		err error
	)
	if d.tx != nil {
		_, err = d.tx.ExecContext(ctx, sqlSyntax, param...)
	} else {
		_, err = d.data.ExecContext(ctx, sqlSyntax, param...)
	}
	return err
}

func (d *Datasource) EndTx(ctx context.Context, errs ...error) error {
	if d.tx == nil {
		return nil
	}
	for _, v := range errs {
		if v != nil {
			d.tx.Rollback()
			return v
		}
	}
	d.tx.Commit()
	return nil
}
