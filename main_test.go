package main

import (
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func testClient(handler http.Handler) (*http.Client, func()) {
	server := httptest.NewServer(handler)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, server.Listener.Addr().String())
			},
		},
	}
	return client, server.Close
}

type mockHTTPResponse struct {
	method string
	URL    string
	body   string
}

func (r mockHTTPResponse) String() string {
	return r.method + " " + r.URL
}

func getResponse(responses []mockHTTPResponse, method string, url string) mockHTTPResponse {
	for _, res := range responses {
		if res.method == method && res.URL == url {
			return res
		}
	}

	return mockHTTPResponse{}
}

func TestGetDays(t *testing.T) {
	days := getDays("")

	if days != 1 {
		t.Errorf("Expected default days to equal 1")
	}

	days = getDays("7")

	if days != 7 {
		t.Errorf("Expected parsed days to equal 7")
	}
}

func TestDryRun(t *testing.T) {
	ex := NewExecutor("token", true)
	db, _ := ex.deleteBranches("user", "repo", []string{"branch"})
	if db != 0 {
		t.Errorf("Expected no branches to be deleted")
	}
}

//todo: test a full third page to make sure the function doesn't keep iterating through pages of old pull requests
func TestListClosedPullRequests(t *testing.T) {
	now := time.Now()

	prs := make([]pullRequest, 100)

	for i := range prs {
		prs[i].Number = uint32(i) + 1
		prs[i].UpdatedAt = now
	}

	firstPageJSON, _ := json.Marshal(prs)
	firstPageResponse := mockHTTPResponse{
		method: "GET",
		URL:    "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=1",
		body:   string(firstPageJSON),
	}

	for i := range prs {
		prs[i].Number += +100
		prs[i].UpdatedAt = now.AddDate(0, 0, 0-(i+1))
	}

	secondPageJSON, _ := json.Marshal(prs)
	secondPageResponse := mockHTTPResponse{
		method: "GET",
		URL:    "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=2",
		body:   string(secondPageJSON),
	}

	thirdPageResponse := mockHTTPResponse{
		method: "GET",
		URL:    "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=3",
		body:   `[]`,
	}

	responses := []mockHTTPResponse{firstPageResponse, secondPageResponse, thirdPageResponse}

	//todo: DRY this
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := getResponse(responses, r.Method, r.URL.String())

		if res.method == "" {
			t.Fatalf(r.URL.String())
		}

		_, err := w.Write([]byte(res.body))

		if err != nil {
			t.Errorf(err.Error())
		}
	})
	client, teardown := testClient(handler)
	defer teardown()

	ex := NewExecutor("token", false)
	ex.client = client
	ex.http = true

	pullRequests, _ := ex.listClosedPullRequests("user", "repo", 20)

	if len(pullRequests) != 120 {
		t.Errorf("Expected only 120 pull requests, got " + strconv.Itoa(len(pullRequests)))
	}
}

func TestRun(t *testing.T) {
	now := time.Now().UTC()

	expectedRequests := make([]string, 0, 1)

	responses := []mockHTTPResponse{
		mockHTTPResponse{
			method: "GET",
			URL:    "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=1",
			body: `[
				{
					"number": 1,
					"updated_at": "` + now.Format(time.RFC3339) + `",
					"head": {
						"ref": "stalebranch",
						"sha": "1761e021e70d29619ca270046b23bd243f652b98"
					}
				}
				]`,
		},
		mockHTTPResponse{
			method: "GET",
			URL:    "/repos/user/repo/branches?protected=false&per_page=100&page=1",
			body: `[
				{
					"name": "stalebranch",
					"commit": {
						"sha": "1761e021e70d29619ca270046b23bd243f652b98"
					}
				}
				]`,
		},
		mockHTTPResponse{
			method: "DELETE",
			URL:    "/repos/user/repo/git/refs/heads/stalebranch",
			body:   "",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := getResponse(responses, r.Method, r.URL.String())

		if res.method == "" {
			t.Fatalf(r.URL.String())
		}

		expectedRequests = append(expectedRequests, res.String())

		_, err := w.Write([]byte(res.body))

		if err != nil {
			t.Errorf(err.Error())
		}
	})
	client, teardown := testClient(handler)
	defer teardown()

	ex := NewExecutor("token", false)
	ex.client = client
	ex.http = true

	err := Run("user", "repo", 1, *ex)

	if err != nil {
		t.Errorf(err.Error())
	}

	for _, r := range responses {
		found := false
		for _, er := range expectedRequests {
			if r.String() == er {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected request: " + r.String())
		}
	}

	err = Run("", "repo", 1, *ex)

	if err == nil {
		t.Errorf("Expected error on missing user")
	}

	err = Run("user", "", 1, *ex)

	if err == nil {
		t.Errorf("Expected error on missing repo")
	}

	err = Run("user", "repo", 0, *ex)

	if err == nil {
		t.Errorf("Expected error on invalid days")
	}
}

func TestUsage(t *testing.T) {
	setupUsage()
	flag.Usage()
}

func TestMainFn(t *testing.T) {
	_crash := crash

	defer func() { crash = _crash }()

	crash = func(msg string, v ...interface{}) {
		if msg == "" {
			t.Errorf("Expected error")
		}
	}

	main()
}
