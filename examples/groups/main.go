package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/lars"
)

func main() {

	l := lars.New()
	l.Use(Logger)

	l.Get("/", Home)

	users := l.Group("/users")
	users.Get("", Users)
	users.Get("/:id", User)
	users.Get("/:id/profile", UserProfile)

	admins := l.Group("/admins")
	admins.Get("", Admins)
	admins.Get("/:id", Admin)
	admins.Get("/:id/profile", AdminProfile)

	http.ListenAndServe(":3007", l.Serve())
}

// Home ...
func Home(c *lars.Context) {
	c.Response.Write([]byte("Welcome Home"))
}

// Users ...
func Users(c *lars.Context) {
	c.Response.Write([]byte("Users"))
}

// User ...
func User(c *lars.Context) {
	c.Response.Write([]byte("User"))
}

// UserProfile ...
func UserProfile(c *lars.Context) {
	c.Response.Write([]byte("User Profile"))
}

// Admins ...
func Admins(c *lars.Context) {
	c.Response.Write([]byte("Admins"))
}

// Admin ...
func Admin(c *lars.Context) {
	c.Response.Write([]byte("Admin"))
}

// AdminProfile ...
func AdminProfile(c *lars.Context) {
	c.Response.Write([]byte("Admin Profile"))
}

// Logger ...
func Logger(c *lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request.URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request.Method, c.Response.Status(), path, stop.Sub(start))
}
