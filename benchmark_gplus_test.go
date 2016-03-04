package lars

import (
	"net/http"
	"testing"
)

// Google+
// https://developers.google.com/+/api/latest/
// (in reality this is just a subset of a much larger API)
var gplusAPI = []route{
	// People
	{"GET", "/people/:userId"},
	{"GET", "/people"},
	{"GET", "/activities/:activityId/people/:collection"},
	{"GET", "/people/:userId/people/:collection"},
	{"GET", "/people/:userId/openIdConnect"},

	// Activities
	{"GET", "/people/:userId/activities/:collection"},
	{"GET", "/activities/:activityId"},
	{"GET", "/activities"},

	// Comments
	{"GET", "/activities/:activityId/comments"},
	{"GET", "/comments/:commentId"},

	// Moments
	{"POST", "/people/:userId/moments/:collection"},
	{"GET", "/people/:userId/moments/:collection"},
	{"DELETE", "/moments/:id"},
}

var gplusLARS http.Handler

func init() {
	calcMem("gplusAPI", func() {
		gplusLARS = loadLARS(gplusAPI)
	})

}

func BenchmarkLARS_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusLARS, req)
}

func BenchmarkLARS_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusLARS, req)
}

func BenchmarkLARS_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusLARS, req)
}

func BenchmarkLARS_GPlusAll(b *testing.B) {
	benchRoutes(b, gplusLARS, gplusAPI)
}
