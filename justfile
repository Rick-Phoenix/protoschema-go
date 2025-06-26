set dotenv-load := true

[working-directory("proto")]
generate:
    buf generate

[working-directory("proto")]
buf-deps:
  buf dep update

dbtest:
    go run ./cmd/dbtest/main.go

protogen:
    go run ./cmd/protogen/main.go
