// cmd/server/main.go
package main

import (
	"log"
	"net/http" // Still needed for http.Handler and status codes, etc.

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware" // For standard middleware like Logger, Recover
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	greetv1connect "myapi/gen/greet/v1/greetv1connect"
	"myapi/internal/handler"
)

func main() {
	log.Println("🚀 Starting server with Echo...")

	// 1. Initialize your GreetServer (same as before)
	greetServer, err := handler.NewGreetServer()
	if err != nil {
		log.Fatalf("🚨 Failed to create greet server: %v", err)
	}
	log.Println("✅ GreetServer initialized")

	// 2. Create ConnectRPC handler (same as before)
	// path is the base path for the service (e.g., "/greet.v1.GreetService/")
	// connectHandler is an http.Handler
	path, connectHandler := greetv1connect.NewGreetServiceHandler(greetServer)
	log.Printf("🔗 ConnectRPC handler ready for path prefix: %s", path)

	// 3. Create a new Echo instance
	e := echo.New()

	// 4. Add standard Echo middleware (optional, but recommended)
	e.Use(middleware.Logger())  // Logs HTTP requests
	e.Use(middleware.Recover()) // Recovers from panics and sends a 500

	// 5. Mount the ConnectRPC handler onto Echo
	// We need to handle any requests that start with the 'path' prefix.
	// Echo's routing for "any" method under a path group is a good way.
	// The `*` in the route path means it will match any sub-path.
	// For example, if path is "/greet.v1.GreetService/", this will match:
	// - /greet.v1.GreetService/Greet
	// - /greet.v1.GreetService/AnotherMethod
	//
	// The connectHandler itself will further route based on the full path.
	e.Any(path+"*", func(c echo.Context) error {
		// Let the original ConnectRPC handler do its thing
		// We pass Echo's ResponseWriter and Request to it.
		connectHandler.ServeHTTP(c.Response().Writer, c.Request())
		return nil // Return nil as ServeHTTP handles the response
	})
	log.Printf("🛣️ Echo will route requests starting with %s* to ConnectRPC handler", path)

	// You can add other Echo routes here if needed for REST APIs, static files, etc.
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Echo! Your ConnectRPC service is also running.")
	})

	// 6. Define the server address (same as before)
	address := "localhost:8080"
	log.Printf("👂 Server will listen on %s", address)

	// 7. Start the Echo server with H2C support
	// Echo's e.Start() doesn't directly support h2c.
	// So, we create an http.Server like before, but give Echo's instance
	// as the handler (wrapped in h2c).
	httpServer := &http.Server{
		Addr:    address,
		Handler: h2c.NewHandler(e, &http2.Server{}), // Pass Echo instance 'e' to h2c
	}

	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("💥 Echo server failed to listen and serve: %v", err)
	}
}
