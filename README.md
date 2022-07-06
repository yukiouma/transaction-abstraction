# Transaction Abstraction

A demo for transaction abstraction in repository layer adapting different orm or driver.



## Abstraction in repository

repository definition:

```go
type Repository interface {
	StartTx(ctx context.Context) (Repository, error)
	EndTx(ctx context.Context, errs ...error) error
}
```

### StartTx

Return a new Repository object with transaction mode.

### EndTx

Commit or rollback it self, according valid errors passed into parameters, if there is at least one valid error, the transaction will call back, otherwise it will commit.



## Use transaction in domain or usecase

### example

Assumption

> Suppose we want to create an admin user in MySQL, create a record in the table users, and create a record containing the user id and role id in the table user-roles (assuming admin's role id is 1).

#### Repository

```go
type Repository interface {
	StartTx(ctx context.Context) (Repository, error)
	CreateUser(ctx context.Context, name ...string) ([]int, error)
	CreateUserRole(ctx context.Context, data ...*valueobject.UserRole) error
	EndTx(ctx context.Context, errs ...error) error
}
```

#### Usecase 

```go
type Usecase struct {
	repo repository.Repository
}
```

#### Method Definition

```go
func (u *Usecase) CreateAdminUser(ctx context.Context, name ...string) error {
    // initialize a transaction repository first
	tx, err := u.repo.StartTx(ctx)
	if err != nil {
		return err
	}
    // operating repository by transaction repo
	newUserId, err1 := tx.CreateUser(ctx, name...)
	if err1 != nil {
        // end transaction when meets error
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
        // end transaction when meets error
		tx.EndTx(ctx, err2)
		return err2
	}
    // commit transaction, then transaction repository will be release with the end of method
	return tx.EndTx(ctx)
}
```



## Repository Implements



### entgo

[Reference of entgo](https://entgo.io/docs/getting-started)



implement in datasource:

```go
type Datasource struct {
	client *ent.Client
	tx     *ent.Tx
}

var _ repository.Repository = (*Datasource)(nil)

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error)

func (d *Datasource) EndTx(ctx context.Context, errs ...error) error

```

Struct `Datasource` is the implement of interface `Repository`.Since  `ent.Client` and `ent.Tx` does not implement the same interface, so we store pointer of `ent.Tx` into `Datasource`



#### Begin transaction

```go
func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	tx, err := d.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &Datasource{client: d.client, tx: tx}, nil
}
```

Refer to [usage of transaction in ent](https://entgo.io/docs/translations), we should create a transaction session from client first, then operate database with this session, then initialize `ent.Tx` in method `StartTx`，and return a new `Datasource` struct to isolate different transaction sessions.



#### Commit or rollback

```go
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
```

count the valid errors, rollback the transaction if a valid error is encountered, otherwise commit the transaction.





#### implement methods in repository

implement the methods defined in repository in [example](#example).

```go
func (d *Datasource) CreateUser(ctx context.Context, name ...string) ([]int, error) {
	var executor *ent.UserClient
    // to determine if use transaction or not
	if d.tx != nil {
		executor = d.tx.User
	} else {
		executor = d.client.User
	}
    // bulk create user and get new user id
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
```

Choose general client or transaction client first, the rest of operation is the same in different client mode. The timing of start transaction, commit or rollback is handed over to usecase layer.



### gorm

[Reference of gorm](https://gorm.io/docs/index.html)



implement in datasource:

```go
type Datasource struct {
	data *gorm.DB
}

var _ repository.Repository = (*Datasource)(nil)

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error)

func (d *Datasource) EndTx(ctx context.Context, errs ...error) error

```

Struct `Datasource` is the implement of interface `Repository`.



#### Begin transaction

```go
func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	return NewDatasource(d.data.Begin()), nil
}
```

Refer to [usage of manual transaction in gorm](https://gorm.io/docs/transactions.html#Control-the-transaction-manually), we should initialize a `*gorm.DB` object with transaction using the original `*gorm.DB`(will return a `*gorm.DB` which is clone from the original one), and initialize a new `Datasource` object using the new `*gorm.DB`.



#### Commit or rollback

```go
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

```

count the valid errors, rollback the transaction if a valid error is encountered, otherwise commit the transaction.



#### implement methods in repository

implement the methods defined in repository in [example](#example).

```go
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
```

There is no difference no matter using transaction or not when create, update or delete data





### go-sql-driver



implement in datasource:

```go
type Datasource struct {
	data *sql.DB
	tx   *sql.Tx
}

var _ repository.Repository = (*Datasource)(nil)

func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error)

func (d *Datasource) EndTx(ctx context.Context, errs ...error) error

```

Struct `Datasource` is the implement of interface `Repository`.Since  `sql.DB` and `sql.Tx` does not implement the same interface, so we store pointer of `sql.Tx` into `Datasource`



#### Begin transaction

```go
func (d *Datasource) StartTx(ctx context.Context) (repository.Repository, error) {
	tx, err := d.data.Begin()
	if err != nil {
		return nil, err
	}
	ds := NewDatasource(d.data)
	ds.tx = tx
	return ds, nil
}
```

We should create a transaction session from client first, then operate database with this session, then initialize `sql.Tx` in method `StartTx`，and return a new `Datasource` struct to isolate different transaction sessions.



#### Commit or rollback

```go
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

```

count the valid errors, rollback the transaction if a valid error is encountered, otherwise commit the transaction.



#### implement methods in repository

implement the methods defined in repository in [example](#example).

```go
func (d *Datasource) CreateUser(ctx context.Context, name ...string) ([]int, error) {

	// build syntax
	// ignore....

	// select execution mode
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
	// execute creation
    // ignore....
}
```

Choose general client or transaction client first, the rest of operation is the same in different client mode. The timing of start transaction, commit or rollback is handed over to usecase layer.