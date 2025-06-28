package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	_ "embed"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/Rick-Phoenix/gofirst/db"
	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
	"github.com/Rick-Phoenix/gofirst/gen/converter"
	"github.com/Rick-Phoenix/gofirst/gen/myappv1"
	"github.com/Rick-Phoenix/gofirst/gen/myappv1/myappv1connect"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type ctxKey int

const (
	AppCtxKey ctxKey = iota
)

type App struct {
	Store       *db.Store
	UserService *UserService
}

func NewApp() *App {
	database, err := sql.Open("sqlite", "db/database.sqlite3?_time_format=sqlite")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	store := db.NewStore(database)
	userService := NewUserService(store)

	return &App{
		Store: store, UserService: userService,
	}
}

type UserService struct {
	Store *db.Store
}

func NewUserService(s *db.Store) *UserService {
	return &UserService{Store: s}
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {
	userID := req.Msg.GetId()

	user, err := s.Store.GetUserWithPosts(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				switch sqliteErr.Code() {
				case sqlite3.SQLITE_CONSTRAINT:
					//
				}
			}
		}
	}

	convUser := converter.UserToUserMsg(user)

	return connect.NewResponse(&myappv1.GetUserResponse{
		User: convUser,
	}), nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[myappv1.UpdateUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID := req.Msg.GetId()

	err := s.Store.Queries.UpdateUser(ctx, sqlgen.UpdateUserParams{
		Name: "gianfranchino", Id: userID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func main() {
	e := echo.New()

	dbConn, err := sql.Open("sqlite", "db/database.sqlite3?_time_format=sqlite")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	store := db.NewStore(dbConn)

	interceptor, err := validate.NewInterceptor()
	if err != nil {
		log.Fatalf("Error while creating the interceptor: %v", err)
	}

	userService := NewUserService(store)

	mux := http.NewServeMux()
	path, handler := myappv1connect.NewUserServiceHandler(userService, connect.WithInterceptors(interceptor))
	mux.Handle(path, handler)

	e.Any("/*", echo.WrapHandler(mux))

	log.Println("Starting Echo server on :8080 (serving ConnectRPC and other HTTP routes)")
	e.Logger.Fatal(e.Start(":8080"))
}
