package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// todo: get pull requests (paginated) using ?head=BRANCH for every open unprotected BRANCH, check if they're merged
// todo: Dockerfile, GitHub action
// todo: long flags, flags and env vars for --token, --repo, optional --owner (implicitly check all repos)
// todo: do not require token so public repos can be analyzed

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/OWNER/REPO/branches?protected=false&per_page=100", nil)
	req.Header.Add("Authorization", "token "+token)
	if err != nil {
		panic(err)
	}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Printf("%s", body)
}
