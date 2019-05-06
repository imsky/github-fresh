class GithubFresh < Formula
  desc "Keep your GitHub repo fresh"
  homepage "https://github.com/imsky/github-fresh"
  head "https://github.com/imsky/github-fresh", :using => :git, :tag => "v0.9.0"

  depends_on "go" => :build

  def install
    # create github-fresh-darwin in the build directory
    system "make", "build-darwin"
    # rename the produced binary to the canonical name
    mv "github-fresh-darwin", "github-fresh"
    # move github-fresh binary to the right homebrew path
    bin.install "github-fresh"
  end

  test do
    system "false"
  end
end
