ENV = DEV

ent:
	export ENV=$(ENV) && go run ./adapt-ent/cmd/