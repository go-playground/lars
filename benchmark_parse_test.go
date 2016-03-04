package lars

import (
	"net/http"
	"testing"
)

var parseAPI = []route{
	// Objects
	{"POST", "/1/classes/:className"},
	{"GET", "/1/classes/:className/:objectId"},
	{"PUT", "/1/classes/:className/:objectId"},
	{"GET", "/1/classes/:className"},
	{"DELETE", "/1/classes/:className/:objectId"},

	// Users
	{"POST", "/1/users"},
	{"GET", "/1/login"},
	{"GET", "/1/users/:objectId"},
	{"PUT", "/1/users/:objectId"},
	{"GET", "/1/users"},
	{"DELETE", "/1/users/:objectId"},
	{"POST", "/1/requestPasswordReset"},

	// Roles
	{"POST", "/1/roles"},
	{"GET", "/1/roles/:objectId"},
	{"PUT", "/1/roles/:objectId"},
	{"GET", "/1/roles"},
	{"DELETE", "/1/roles/:objectId"},

	// Files
	{"POST", "/1/files/:fileName"},

	// Analytics
	{"POST", "/1/events/:eventName"},

	// Push Notifications
	{"POST", "/1/push"},

	// Installations
	{"POST", "/1/installations"},
	{"GET", "/1/installations/:objectId"},
	{"PUT", "/1/installations/:objectId"},
	{"GET", "/1/installations"},
	{"DELETE", "/1/installations/:objectId"},

	// Cloud
	// Functions
	{"POST", "/1/functions"},
}
var parseLARS http.Handler

func init() {
	calcMem("parseAPI", func() {
		parseLARS = loadLARS(parseAPI)
	})

}

func BenchmarkLARS_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseLARS, req)
}

func BenchmarkLARS_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseLARS, req)
}

func BenchmarkLARS_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseLARS, req)
}

func BenchmarkLARS_ParseAll(b *testing.B) {
	benchRoutes(b, parseLARS, parseAPI)
}
