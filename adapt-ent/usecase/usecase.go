package usecase

import (
	"context"

	"github.com/AkiOuma/transaction-abstraction/adapt-ent/domain/repository"
	"github.com/AkiOuma/transaction-abstraction/adapt-ent/domain/valueobject"
)

type Usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) CreateAdminUser(ctx context.Context, name ...string) error {
	// errs := make([]error, 0, 2)
	tx, err := u.repo.StartTx(ctx)
	if err != nil {
		return err
	}
	newUserId, err1 := tx.CreateUser(ctx, name...)
	if err1 != nil {
		tx.EndTx(ctx, err1)
		return err1
	}
	userrole := make([]*valueobject.UserRole, 0, len(newUserId))
	for _, v := range newUserId {
		userrole = append(userrole, &valueobject.UserRole{
			UserId: v,
			RoleId: []int{1},
		})
	}
	err2 := tx.CreateUserRole(ctx, userrole...)
	if err2 != nil {
		tx.EndTx(ctx, err2)
		return err2
	}
	return tx.EndTx(ctx)
}

func (u *Usecase) CreateUser(ctx context.Context, name ...string) error {
	_, err := u.repo.CreateUser(ctx, name...)
	return err
}
