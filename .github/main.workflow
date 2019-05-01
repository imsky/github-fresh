workflow "github-fresh" {
  on = "pull_request"
  resolves = ["github-fresh"]
}

action "github-fresh" {
  uses = "imsky/github-fresh@v0.7.0"
  secrets = ["GITHUB_TOKEN"]
}
