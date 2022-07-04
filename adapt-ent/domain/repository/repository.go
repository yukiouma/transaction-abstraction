package repository

import (
	"context"

	"github.com/AkiOuma/transaction-abstraction/adapt-ent/domain/valueobject"
)

type Repository interface {
	StartTx(ctx context.Context) (Repository, error)
	CreateUser(ctx context.Context, name ...string) ([]int, error)
	CreateUserRole(ctx context.Context, data ...*valueobject.UserRole) error
	EndTx(ctx context.Context, errs ...error) error
}
