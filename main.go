package main

import (
	"fmt"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = make(map[string]User) // mock database

func createUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	users[user.ID] = user
	c.JSON(http.StatusOK, gin.H{"status": "created"})
}

func getUser(c *gin.Context) {
	userID := c.Param("id")
	user, ok := users[userID]
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
	userID := c.Param("id")
	user, ok := users[userID]
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var updateUser User
	if err := c.BindJSON(&updateUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.Name = updateUser.Name
	user.Email = updateUser.Email
	users[userID] = user
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func deleteUser(c *gin.Context) {
	userID := c.Param("id")
	if _, ok := users[userID]; !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	delete(users, userID)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func tracer() *trace.TracerProvider {
	// Initialize the tracer
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return tp
}

func main() {
	r := gin.Default()

	tp := tracer()
	otel.SetTracerProvider(tp)

	// Use the OpenTelemetry middleware for Gin
	r.Use(otelgin.Middleware("my-api"))

	// Define CRUD routes
	r.POST("/users", createUser)
	r.GET("/users/:id", getUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)

	fmt.Println("Server is running on port 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
