package datasource

import (
	"context"
	"log"
	"os"

	"github.com/AkiOuma/transaction-abstraction/adapt-ent/datasource/ent"
	"github.com/AkiOuma/transaction-abstraction/adapt-ent/datasource/ent/migrate"
	"github.com/AkiOuma/transaction-abstraction/adapt-ent/domain/repository"
	"github.com/AkiOuma/transaction-abstraction/adapt-ent/domain/valueobject"
	_ "github.com/go-sql-driver/mysql"
)

type Datasource struct {
	client *ent.Client
	tx     *ent.Tx
}

var _ repository.Repository = (*Datasource)(nil)

func NewDatasource(client *ent.Client) repository.Repository {
	return &Datasource{
		client: client,
	}
}

func NewEnt(addr string) *ent.Client {
	client, err := ent.Open("mysql", addr)
	if err != nil {
		log.Fatal("ent client init failed, reason: " + err.Error())
	}
	err = client.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		log.Fatal("ent schema init failed, reason: " + err.Error())
	}
	if os.Getenv("ENV") == "DEV" {
		client = client.Debug()
	}
	return client
}

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	tx, err := d.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &Datasource{client: d.client, tx: tx}, nil
}

func (d *Datasource) CreateUser(ctx context.Context, name ...string) ([]int, error) {
	var executor *ent.UserClient
	if d.tx != nil {
		executor = d.tx.User
	} else {
		executor = d.client.User
	}
	creation := make([]*ent.UserCreate, 0, len(name))
	for _, v := range name {
		creation = append(creation, executor.Create().SetName(v))
	}
	result, err := executor.CreateBulk(creation...).Save(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]int, 0, len(result))
	for _, v := range result {
		users = append(users, v.ID)
	}
	return users, nil
}

func (d *Datasource) CreateUserRole(ctx context.Context, data ...*valueobject.UserRole) error {
	var executor *ent.UserRoleClient
	if d.tx != nil {
		executor = d.tx.UserRole
	} else {
		executor = d.client.UserRole
	}
	creation := make([]*ent.UserRoleCreate, 0, len(data)*2)
	for _, user := range data {
		for _, role := range user.RoleId {
			creation = append(creation, executor.Create().SetUserID(user.UserId).SetRoleID(role))
		}
	}
	_, err := executor.CreateBulk(creation...).Save(ctx)
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
	return d.tx.Commit()
}
