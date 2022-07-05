ENV = DEV

ent:
	export ENV=$(ENV) && go run ./adapt-ent/cmd/

gorm:
	export ENV=$(ENV) && go run ./adapt-gorm/cmd/

sql:
	export ENV=$(ENV) && go run ./adapt-sql-driver/cmd/