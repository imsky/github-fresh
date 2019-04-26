class GithubFresh < Formula
  desc "Keep your GitHub repo fresh"
  homepage "https://github.com/imsky/github-fresh"
  url "https://github.com/imsky/github-fresh/archive/v0.7.0.tar.gz"
  sha256 "c4c68cb8d4a906f2882f6abd6ed7525189f1edc55156b6a5641876b2fe2ce39d"
  depends_on "go" => :build

  def install
    system "make", "build-darwin"
  end

  test do
    system "false"
  end
end
