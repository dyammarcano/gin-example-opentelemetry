package main

import (
	"fmt"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	exporter "go.opentelemetry.io/otel/exporters/otlp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = make(map[string]User) // mock database

func createUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("createUserHandler")

	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, span := tracer.Start(ctx, "createUser")
	defer span.End()

	users[user.ID] = user
	c.JSON(http.StatusOK, user)
}

func getUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("getUserHandler")

	userID := c.Param("id")
	ctx, span := tracer.Start(ctx, "getUser")
	defer span.End()

	user, ok := users[userID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("updateUserHandler")

	userID := c.Param("id")
	ctx, span := tracer.Start(ctx, "updateUser")
	defer span.End()

	var userUpdates User
	if err := c.BindJSON(&userUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := users[userID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	users[userID] = User{
		ID:    userID,
		Name:  userUpdates.Name,
		Email: userUpdates.Email,
	}

	c.JSON(http.StatusOK, users[userID])
}

func deleteUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("deleteUserHandler")

	userID := c.Param("id")
	ctx, span := tracer.Start(ctx, "deleteUser")
	defer span.End()

	if _, ok := users[userID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	delete(users, userID)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func listUseresHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("listUseresHandler")

	ctx, span := tracer.Start(ctx, "listUseres")
	defer span.End()

	c.JSON(http.StatusOK, users)
}

func tracer() (tp *trace.TracerProvider) {
	// Initialize the tracer
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	return
}

func main() {
	r := gin.Default()
	otel.SetTracerProvider(tracer())

	// Use the OpenTelemetry middleware for Gin
	r.Use(otelgin.Middleware("my-api"))

	// Define CRUD routes
	r.Handle(http.MethodPost, "/users", createUserHandler)
	r.Handle(http.MethodGet, "/users/:id", getUserHandler)
	r.Handle(http.MethodPut, "/users/:id", updateUserHandler)
	r.Handle(http.MethodGet, "/users", listUseresHandler)
	r.Handle(http.MethodDelete, "/users/:id", deleteUserHandler)

	fmt.Println("Server is running on port 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
