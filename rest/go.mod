module github.com/LeperGnome/simple-chat/rest

go 1.22.2

require github.com/LeperGnome/simple-chat/pkg/chronicler v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)

replace github.com/LeperGnome/simple-chat/pkg/chronicler => ./../pkg/chronicler
