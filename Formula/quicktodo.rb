class Quicktodo < Formula
  desc "Simple todo list CLI built for AI coding workflows"
  homepage "https://github.com/justingosan/quicktodo"
  url "https://github.com/justingosan/quicktodo/archive/v1.0.0.tar.gz"
  sha256 "" # You'll need to fill this in after uploading the release
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    output = shell_output("#{bin}/quicktodo version")
    assert_match "QuickTodo CLI v", output
  end
end