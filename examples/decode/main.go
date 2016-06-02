package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/lars"
)

const (
	// used for io.LimitReader(...) for safetly!
	maxBytes = 16 << 10
)

// User ...
type User struct {
	ID   int `form:"id"`
	Name string
}

func main() {

	l := lars.New()
	l.Post("/user/:id", PostUser)

	go simulateFormPost()

	http.ListenAndServe(":3007", l.Serve())
}

// PostUser ...
func PostUser(c lars.Context) {
	var user User

	// first argument denotes yes or no I would like URL query parameter fields
	// to be included. i.e. 'id' in route '/user/:id' should it be included.
	// run, then change to false and you'll see user.ID is not populated.
	if err := c.Decode(true, maxBytes, &user); err != nil {
		log.Println(err)
	}

	log.Printf("%#v", user)
}

func simulateFormPost() {
	time.Sleep(1000)
	http.PostForm("http://localhost:3007/user/13", url.Values{"Name": {"joeybloggs"}})
}
