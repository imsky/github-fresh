# github-fresh

[![pipeline status](https://gitlab.com/imsky/github-fresh/badges/master/pipeline.svg)](https://gitlab.com/imsky/github-fresh/commits/master) [![coverage report](https://gitlab.com/imsky/github-fresh/badges/master/coverage.svg)](https://gitlab.com/imsky/github-fresh/commits/master)

github-fresh deletes branches of closed pull requests.

## Usage

```bash
# default usage
$ github-fresh -user=imsky -repo=github-fresh

# see what branches would be deleted from pull requests closed in the last month
$ github-fresh -user=imsky -repo=github-fresh -days=30 -dry=true

$ github-fresh -help
  -days int
    	Max age in days of checked pull requests (GITHUB_FRESH_DAYS) (default 1)
  -dry
    	Dry run (GITHUB_FRESH_DRY)
  -repo string
    	GitHub repo (GITHUB_FRESH_REPO)
  -token string
    	GitHub API token (GITHUB_FRESH_TOKEN)
  -user string
    	GitHub user (GITHUB_FRESH_USER)
```

## Installation

### Homebrew

### Releases

Download the binary for your platform from the [releases](https://github.com/imsky/github-fresh/releases) page.

### Docker

```sh
$ docker pull imsky/github-fresh
$ docker run -it -e GITHUB_FRESH_TOKEN imsky/github-fresh -user=imsky -repo=github-fresh -dry=true
```

### Go

```sh
$ go get -u github.com/imsky/github-fresh
```

## License

github-fresh is provided under the [MIT License](./LICENSE).

## Credits

github-fresh is a project by [Ivan Malopinsky](http://imsky.co).
