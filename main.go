package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// todo: get closed pull requests sorted by last updated time; stop processing after some cut-off (3h); get all unprotected branches - if any branches match those of the recently closed pull requests, delete them
// todo: long flags, flags and env vars for --token, --repo, optional --owner (implicitly check all repos)
// todo: Dockerfile, GitHub action

type PullRequest struct {
	Number uint64 `json:"number"`

	PullRequestHead struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

type PullRequestResult struct {
	PullRequests []PullRequest
}

func makeGithubAPIRequest(url string, method string) (res *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	token := os.Getenv("GITHUB_TOKEN")
	req.Header.Add("Authorization", "token "+token)
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	res, err := makeGithubAPIRequest("https://api.github.com/repos/OWNER/REPO/pulls?state=closed&sort=updated&direction=desc&per_page=100", "GET")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(res.Body)
	var x PullRequestResult
	err = decoder.Decode(&x.PullRequests)
	fmt.Println(x)
}
