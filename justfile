set dotenv-load := true

[working-directory("proto")]
generate:
    buf generate

dbtest:
    go run ./cmd/dbtest/main.go

protogen:
    go run ./cmd/protogen/main.go

