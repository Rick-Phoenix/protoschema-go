package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// For parsing user ID from params
	"connectrpc.com/connect" // ConnectRPC library
	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"github.com/Rick-Phoenix/gofirst/db"                         // Your database store package
	"github.com/Rick-Phoenix/gofirst/gen/myappv1"                // Generated ConnectRPC interfaces
	"github.com/Rick-Phoenix/gofirst/gen/myappv1/myappv1connect" // Generated ConnectRPC interfaces
	"github.com/Rick-Phoenix/gofirst/gen/protoconvert"
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

// GetUser implements the GetUser RPC from your proto.
func (s *UserService) GetUser(
	ctx context.Context, // The context from connectrpc
	req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {
	userID := req.Msg.GetId()

	user, err := s.Store.GetUserWithPosts(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	convUser := protoconvert.UserToUserMsg(user)
	return connect.NewResponse(&myappv1.GetUserResponse{
		User: convUser,
	}), nil
}

func main() {
	e := echo.New()

	// 1. Initialize your database connection
	// Replace with your actual SQLite path (e.g., in a data directory)
	dbConn, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	// 2. Create your Store instance
	store := db.NewStore(dbConn)

	// 3. Apply the StoreMiddleware globally
	e.Use(StoreMiddleware(store))

	// 4. Create your ConnectRPC server mux
	mux := http.NewServeMux()
	// Register your MyService implementation with the ConnectRPC mux
	path, handler := myappv1connect.NewMyServiceHandler(NewMyServiceServer())
	mux.Handle(path, handler)

	// 5. Register the ConnectRPC handler with Echo
	// This will route all requests starting with the ConnectRPC service path to the mux.
	e.Any("/*", echo.WrapHandler(mux)) // Catches all paths, use more specific route if preferred

	// Optional: Add a simple health check or non-RPC route
	e.GET("/health", func(c echo.Context) error {
		// You can still access the CustomContext here too
		cc, ok := c.(*AppCtx)
		if !ok {
			return c.String(http.StatusInternalServerError, "Context not custom")
		}
		// Example: Check DB health via store
		err := cc.Store.db.(*sql.DB).PingContext(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("DB down: %v", err))
		}
		return c.String(http.StatusOK, "API Healthy! DB Ping OK.")
	})

	// Start the server
	log.Println("Starting Echo server on :8080 (serving ConnectRPC and other HTTP routes)")
	e.Logger.Fatal(e.Start(":8080"))
}
