package lars

import (
	"net/http"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

// func TestMain(m *testing.M) {
// 	flag.Parse()
// 	os.Exit(m.Run())
// }

func TestFind(t *testing.T) {
	l := New()

	// fn := []Handler{func(c Context) {
	// 	c.Response().Write([]byte(c.Params()[0].Value))
	// }}

	// for _, r := range githubAPI {
	// 	l.RouteGroup.handle(r.method, r.path, fn)
	// }

	// l.Delete("/authorizations/:id", func(c Context) {

	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })
	// l.Get("/test/two/three/", func(c Context) { c.Response().Write([]byte("in three")) })
	// l.Get("/test/two/three", func(Context) {})
	// l.Get("/test/too%2fthree/four", func(Context) {})

	// var body string

	// code, _ := request(GET, "", l)
	// Equal(t, code, http.StatusNotFound)
	//

	// l.Get("/authorizations", func(c Context) {
	// 	// p, _ := c.Param("id")
	// 	// c.Response().Write([]byte(p))
	// })

	// l.Post("/authorizations", func(c Context) {
	// 	// p, _ := c.Param("id")
	// 	// c.Response().Write([]byte(p))
	// })

	// l.Get("/authorizations/:id", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// l.Delete("/authorizations/:id", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// for idx, n := range l.router.tree.static {
	// 	fmt.Println(idx, n.path, n.params == nil, n.chains, n.params.chains)
	// }

	// {"GET", "/authorizations/:id"},
	// {"POST", "/authorizations"},
	// //{"PUT", "/authorizations/clients/:client_id"},
	// //{"PATCH", "/authorizations/:id"},
	// {"DELETE", "/authorizations/:id"},

	// l.Get("/authorizations/:id/test", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// code, body := request(GET, "/authorizations/11/test", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "11")

	l.Get("/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("home"))
	})

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "home")

	l.Get("/authorizations/user/test/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("1"))
	})

	l.Get("/authorizations/:id/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("2"))
	})

	code, body = request(GET, "/authorizations/user/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "2")

	code, body = request(GET, "/authorizations/user/test/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "1")

	// code, body := request(GET, "/authorizations/11", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "11")

	// code, _ = request(GET, "/authorizations", l)
	// Equal(t, code, http.StatusOK)

	// code, _ = request(POST, "/authorizations", l)
	// Equal(t, code, http.StatusOK)

	// code, body = request(DELETE, "/authorizations/13", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "13")

	// r, _ := http.NewRequest("GET", "", nil)
	// w := httptest.NewRecorder()
	// l.serveHTTP(w, r)

	// fmt.Println(l.router.tree.static[0].param.path)
	// fmt.Println(l.router.tree.static[0].params.priority, l.router.tree.static[0].params.static.path)

	// l.router.sort()

	// for idx, n := range l.router.tree.static[0].static {
	// 	fmt.Println(idx, n.priority, n.path)
	// }

	// l.Get("/github.com/go-experimental/lars3/:blob/master历日本語/⌘/à/:alice/*", func(Context) {})
}

var githubAPITest = []route{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	// {"GET", "/applications/:client_id/tokens/:access_token"},
	// {"DELETE", "/applications/:client_id/tokens"},
	// {"DELETE", "/applications/:client_id/tokens/:access_token"},

	// // Activity
	// {"GET", "/events"},
	// {"GET", "/repos/:owner/:repo/events"},
	// {"GET", "/networks/:owner/:repo/events"},
	// {"GET", "/orgs/:org/events"},
	// {"GET", "/users/:user/received_events"},
	// {"GET", "/users/:user/received_events/public"},
	// {"GET", "/users/:user/events"},
	// {"GET", "/users/:user/events/public"},
	// {"GET", "/users/:user/events/orgs/:org"},
	// {"GET", "/feeds"},
	// {"GET", "/notifications"},
	// {"GET", "/repos/:owner/:repo/notifications"},
	// {"PUT", "/notifications"},
	// {"PUT", "/repos/:owner/:repo/notifications"},
	// {"GET", "/notifications/threads/:id"},
	// //{"PATCH", "/notifications/threads/:id"},
	// {"GET", "/notifications/threads/:id/subscription"},
	// {"PUT", "/notifications/threads/:id/subscription"},
	// {"DELETE", "/notifications/threads/:id/subscription"},
	// {"GET", "/repos/:owner/:repo/stargazers"},
	// {"GET", "/users/:user/starred"},
	// {"GET", "/user/starred"},
	// {"GET", "/user/starred/:owner/:repo"},
	// {"PUT", "/user/starred/:owner/:repo"},
	// {"DELETE", "/user/starred/:owner/:repo"},
	// {"GET", "/repos/:owner/:repo/subscribers"},
	// {"GET", "/users/:user/subscriptions"},
	// {"GET", "/user/subscriptions"},
	// {"GET", "/repos/:owner/:repo/subscription"},
	// {"PUT", "/repos/:owner/:repo/subscription"},
	// {"DELETE", "/repos/:owner/:repo/subscription"},
	// {"GET", "/user/subscriptions/:owner/:repo"},
	// {"PUT", "/user/subscriptions/:owner/:repo"},
	// {"DELETE", "/user/subscriptions/:owner/:repo"},

	// // Gists
	// {"GET", "/users/:user/gists"},
	// {"GET", "/gists"},
	// //{"GET", "/gists/public"},
	// //{"GET", "/gists/starred"},
	// {"GET", "/gists/:id"},
	// {"POST", "/gists"},
	// //{"PATCH", "/gists/:id"},
	// {"PUT", "/gists/:id/star"},
	// {"DELETE", "/gists/:id/star"},
	// {"GET", "/gists/:id/star"},
	// {"POST", "/gists/:id/forks"},
	// {"DELETE", "/gists/:id"},

	// // Git Data
	// {"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	// {"POST", "/repos/:owner/:repo/git/blobs"},
	// {"GET", "/repos/:owner/:repo/git/commits/:sha"},
	// {"POST", "/repos/:owner/:repo/git/commits"},
	// //{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	// {"GET", "/repos/:owner/:repo/git/refs"},
	// {"POST", "/repos/:owner/:repo/git/refs"},
	// //{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	// //{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	// {"GET", "/repos/:owner/:repo/git/tags/:sha"},
	// {"POST", "/repos/:owner/:repo/git/tags"},
	// {"GET", "/repos/:owner/:repo/git/trees/:sha"},
	// {"POST", "/repos/:owner/:repo/git/trees"},

	// // Issues
	// {"GET", "/issues"},
	// {"GET", "/user/issues"},
	// {"GET", "/orgs/:org/issues"},
	// {"GET", "/repos/:owner/:repo/issues"},
	// {"GET", "/repos/:owner/:repo/issues/:number"},
	// {"POST", "/repos/:owner/:repo/issues"},
	// //{"PATCH", "/repos/:owner/:repo/issues/:number"},
	// {"GET", "/repos/:owner/:repo/assignees"},
	// {"GET", "/repos/:owner/:repo/assignees/:assignee"},
	// {"GET", "/repos/:owner/:repo/issues/:number/comments"},
	// //{"GET", "/repos/:owner/:repo/issues/comments"},
	// //{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	// {"POST", "/repos/:owner/:repo/issues/:number/comments"},
	// //{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	// //{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	// {"GET", "/repos/:owner/:repo/issues/:number/events"},
	// //{"GET", "/repos/:owner/:repo/issues/events"},
	// //{"GET", "/repos/:owner/:repo/issues/events/:id"},
	// {"GET", "/repos/:owner/:repo/labels"},
	// {"GET", "/repos/:owner/:repo/labels/:name"},
	// {"POST", "/repos/:owner/:repo/labels"},
	// //{"PATCH", "/repos/:owner/:repo/labels/:name"},
	// {"DELETE", "/repos/:owner/:repo/labels/:name"},
	// {"GET", "/repos/:owner/:repo/issues/:number/labels"},
	// {"POST", "/repos/:owner/:repo/issues/:number/labels"},
	// {"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	// {"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	// {"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	// {"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	// {"GET", "/repos/:owner/:repo/milestones"},
	// {"GET", "/repos/:owner/:repo/milestones/:number"},
	// {"POST", "/repos/:owner/:repo/milestones"},
	// //{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	// {"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// // Miscellaneous
	// {"GET", "/emojis"},
	// {"GET", "/gitignore/templates"},
	// {"GET", "/gitignore/templates/:name"},
	// {"POST", "/markdown"},
	// {"POST", "/markdown/raw"},
	// {"GET", "/meta"},
	// {"GET", "/rate_limit"},

	// // Organizations
	// {"GET", "/users/:user/orgs"},
	// {"GET", "/user/orgs"},
	// {"GET", "/orgs/:org"},
	// //{"PATCH", "/orgs/:org"},
	// {"GET", "/orgs/:org/members"},
	// {"GET", "/orgs/:org/members/:user"},
	// {"DELETE", "/orgs/:org/members/:user"},
	// {"GET", "/orgs/:org/public_members"},
	// {"GET", "/orgs/:org/public_members/:user"},
	// {"PUT", "/orgs/:org/public_members/:user"},
	// {"DELETE", "/orgs/:org/public_members/:user"},
	// {"GET", "/orgs/:org/teams"},
	// {"GET", "/teams/:id"},
	// {"POST", "/orgs/:org/teams"},
	// //{"PATCH", "/teams/:id"},
	// {"DELETE", "/teams/:id"},
	// {"GET", "/teams/:id/members"},
	// {"GET", "/teams/:id/members/:user"},
	// {"PUT", "/teams/:id/members/:user"},
	// {"DELETE", "/teams/:id/members/:user"},
	// {"GET", "/teams/:id/repos"},
	// {"GET", "/teams/:id/repos/:owner/:repo"},
	// {"PUT", "/teams/:id/repos/:owner/:repo"},
	// {"DELETE", "/teams/:id/repos/:owner/:repo"},
	// {"GET", "/user/teams"},

	// // Pull Requests
	// {"GET", "/repos/:owner/:repo/pulls"},
	// {"GET", "/repos/:owner/:repo/pulls/:number"},
	// {"POST", "/repos/:owner/:repo/pulls"},
	// //{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	// {"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	// {"GET", "/repos/:owner/:repo/pulls/:number/files"},
	// {"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	// {"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	// {"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	// //{"GET", "/repos/:owner/:repo/pulls/comments"},
	// //{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	// {"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	// //{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	// //{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// // Repositories
	// {"GET", "/user/repos"},
	// {"GET", "/users/:user/repos"},
	// {"GET", "/orgs/:org/repos"},
	// {"GET", "/repositories"},
	// {"POST", "/user/repos"},
	// {"POST", "/orgs/:org/repos"},
	// {"GET", "/repos/:owner/:repo"},
	// //{"PATCH", "/repos/:owner/:repo"},
	// {"GET", "/repos/:owner/:repo/contributors"},
	// {"GET", "/repos/:owner/:repo/languages"},
	// {"GET", "/repos/:owner/:repo/teams"},
	// {"GET", "/repos/:owner/:repo/tags"},
	// {"GET", "/repos/:owner/:repo/branches"},
	// {"GET", "/repos/:owner/:repo/branches/:branch"},
	// {"DELETE", "/repos/:owner/:repo"},
	// {"GET", "/repos/:owner/:repo/collaborators"},
	// {"GET", "/repos/:owner/:repo/collaborators/:user"},
	// {"PUT", "/repos/:owner/:repo/collaborators/:user"},
	// {"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	// {"GET", "/repos/:owner/:repo/comments"},
	// {"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	// {"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	// {"GET", "/repos/:owner/:repo/comments/:id"},
	// //{"PATCH", "/repos/:owner/:repo/comments/:id"},
	// {"DELETE", "/repos/:owner/:repo/comments/:id"},
	// {"GET", "/repos/:owner/:repo/commits"},
	// {"GET", "/repos/:owner/:repo/commits/:sha"},
	// {"GET", "/repos/:owner/:repo/readme"},
	// //{"GET", "/repos/:owner/:repo/contents/*path"},
	// //{"PUT", "/repos/:owner/:repo/contents/*path"},
	// //{"DELETE", "/repos/:owner/:repo/contents/*path"},
	// //{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	// {"GET", "/repos/:owner/:repo/keys"},
	// {"GET", "/repos/:owner/:repo/keys/:id"},
	// {"POST", "/repos/:owner/:repo/keys"},
	// //{"PATCH", "/repos/:owner/:repo/keys/:id"},
	// {"DELETE", "/repos/:owner/:repo/keys/:id"},
	// {"GET", "/repos/:owner/:repo/downloads"},
	// {"GET", "/repos/:owner/:repo/downloads/:id"},
	// {"DELETE", "/repos/:owner/:repo/downloads/:id"},
	// {"GET", "/repos/:owner/:repo/forks"},
	// {"POST", "/repos/:owner/:repo/forks"},
	// {"GET", "/repos/:owner/:repo/hooks"},
	// {"GET", "/repos/:owner/:repo/hooks/:id"},
	// {"POST", "/repos/:owner/:repo/hooks"},
	// //{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	// {"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	// {"DELETE", "/repos/:owner/:repo/hooks/:id"},
	// {"POST", "/repos/:owner/:repo/merges"},
	// {"GET", "/repos/:owner/:repo/releases"},
	// {"GET", "/repos/:owner/:repo/releases/:id"},
	// {"POST", "/repos/:owner/:repo/releases"},
	// //{"PATCH", "/repos/:owner/:repo/releases/:id"},
	// {"DELETE", "/repos/:owner/:repo/releases/:id"},
	// {"GET", "/repos/:owner/:repo/releases/:id/assets"},
	// {"GET", "/repos/:owner/:repo/stats/contributors"},
	// {"GET", "/repos/:owner/:repo/stats/commit_activity"},
	// {"GET", "/repos/:owner/:repo/stats/code_frequency"},
	// {"GET", "/repos/:owner/:repo/stats/participation"},
	// {"GET", "/repos/:owner/:repo/stats/punch_card"},
	// {"GET", "/repos/:owner/:repo/statuses/:ref"},
	// {"POST", "/repos/:owner/:repo/statuses/:ref"},

	// // Search
	// {"GET", "/search/repositories"},
	// {"GET", "/search/code"},
	// {"GET", "/search/issues"},
	// {"GET", "/search/users"},
	// {"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	// {"GET", "/legacy/repos/search/:keyword"},
	// {"GET", "/legacy/user/search/:keyword"},
	// {"GET", "/legacy/user/email/:email"},

	// // Users
	// {"GET", "/users/:user"},
	// {"GET", "/user"},
	// //{"PATCH", "/user"},
	// {"GET", "/users"},
	// {"GET", "/user/emails"},
	// {"POST", "/user/emails"},
	// {"DELETE", "/user/emails"},
	// {"GET", "/users/:user/followers"},
	// {"GET", "/user/followers"},
	// {"GET", "/users/:user/following"},
	// {"GET", "/user/following"},
	// {"GET", "/user/following/:user"},
	// {"GET", "/users/:user/following/:target_user"},
	// {"PUT", "/user/following/:user"},
	// {"DELETE", "/user/following/:user"},
	// {"GET", "/users/:user/keys"},
	// {"GET", "/user/keys"},
	// {"GET", "/user/keys/:id"},
	// {"POST", "/user/keys"},
	// //{"PATCH", "/user/keys/:id"},
	// {"DELETE", "/user/keys/:id"},
}
