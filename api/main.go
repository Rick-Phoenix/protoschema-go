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
)

// --- Custom Echo Context and Helper ---

// CustomContext embeds echo.Context and adds our database Store.
type CustomContext struct {
	echo.Context
	Store *db.Store // Reference to your db.Store
}

// GetCustomContext is a helper function to safely retrieve CustomContext
// from a standard context.Context (like the one connectrpc handlers receive).
func GetCustomContext(ctx context.Context) (*CustomContext, error) {
	echoCtx := ctx.Value(echo.EchoContextKey)
	if echoCtx == nil {
		return nil, fmt.Errorf("echo context not found in standard context")
	}

	cc, ok := echoCtx.(*CustomContext)
	if !ok {
		return nil, fmt.Errorf("failed to assert to CustomContext from echo context")
	}
	return cc, nil
}

// --- Echo Middleware for Store Injection ---

// StoreMiddleware injects the *db.Store instance into the Echo context.
func StoreMiddleware(s *db.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{
				Context: c,
				Store:   s,
			}
			// Call the next handler in the chain with our custom context
			return next(cc)
		}
	}
}

// --- ConnectRPC Service Implementation ---

// MyServiceServer implements the myappv1connect.MyServiceHandler interface.
type MyServiceServer struct {
	// No need to hold the Store here, we get it from the context
}

// NewMyServiceServer creates a new instance of MyServiceServer.
func NewMyServiceServer() *MyServiceServer {
	return &MyServiceServer{}
}

// GetUser implements the GetUser RPC from your proto.
func (s *MyServiceServer) GetUser(
	ctx context.Context, // The context from connectrpc
	req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {
	// 1. Get the CustomContext from the request's HTTP context
	cc, err := GetCustomContext(ctx)
	if err != nil {
		// Log the internal error and return a gRPC-compatible error
		log.Printf("Error getting custom context in GetUser: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to retrieve context"))
	}

	// 2. Access your Store instance
	store := cc.Store
	if store == nil {
		log.Println("Store not found in CustomContext!")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database store not available"))
	}

	// --- Demonstrate using the Store ---
	// Let's assume a user with ID 1 exists for demonstration.
	// In a real app, you might get a user ID from auth info in the context
	// or from a request parameter if this RPC was designed for it.
	demoUserID := int64(1)
	userWithPosts, dbErr := store.GetUserWithPosts(ctx, demoUserID) // Pass the RPC's ctx to DB calls
	if dbErr != nil {
		log.Printf("Failed to fetch user with posts from DB for ID %d: %v", demoUserID, dbErr)
		// Map database error to an appropriate gRPC code
		if dbErr == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user %d not found", demoUserID))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database operation failed: %w", dbErr))
	}

	// If we got here, DB access worked!
	log.Printf("Successfully retrieved user %s from DB via CustomContext in GetUser RPC.", userWithPosts.User.Name)

	// 3. Construct and return the response
	userName := req.Msg.Name
	if userName == "" {
		userName = "World"
	}
	message := fmt.Sprintf("Hello, %s! (DB User retrieved: %s)", userName, userWithPosts.User.Name)

	return connect.NewResponse(&myappv1.GetUserResponse{
		Message: message,
	}), nil
}

// --- Main Application Setup ---

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
		cc, ok := c.(*CustomContext)
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
