package main

import (
	"context"
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/repos/user/repo/pulls?state=closed&sort=updated&direction=desc&per_page=100&page=1" {
			w.Write([]byte(`[
			{
				"number": 1,
				"closed_at": "` + now.Format(time.RFC3339) + `",
				"head": {
					"ref": "stalebranch",
					"sha": "1761e021e70d29619ca270046b23bd243f652b98"
				}
			}
			]`))
		} else if r.URL.String() == "/repos/user/repo/branches?protected=false&per_page=100&page=1" {
			w.Write([]byte(`[
			{
				"name": "stalebranch",
				"commit": {
					"sha": "1761e021e70d29619ca270046b23bd243f652b98"
				}
			}
			]`))
		} else if r.URL.String() == "/repos/user/repo/git/refs/heads/stalebranch" {
			w.Write([]byte(``))
		} else {
			t.Fatalf(r.URL.String())
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
}
