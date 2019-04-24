package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestRun(t *testing.T) {
	now := time.Now().UTC()
	//tooLongAgo := time.Date(2010, 1, 1, 1, 1, 1, 1, time.UTC).UTC()
	responses := []mockHTTPResponse{
		mockHTTPResponse{
			method: "GET",
			URL:    "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=1",
			body: `[
				{
					"number": 1,
					"closed_at": "` + now.Format(time.RFC3339) + `",
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

		_, err := w.Write([]byte(res.body))

		if err != nil {
			t.Errorf(err.Error())
		}
	})
	client, teardown := testClient(handler)
	defer teardown()

	ex := NewExecutor("token")
	ex.client = client
	ex.http = true

	err := Run("user", "repo", 1, *ex)

	if err != nil {
		t.Errorf(err.Error())
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
