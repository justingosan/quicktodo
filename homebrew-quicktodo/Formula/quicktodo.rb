class Quicktodo < Formula
  desc "Simple todo list CLI built for AI coding workflows"
  homepage "https://github.com/justingosan/quicktodo"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/justingosan/quicktodo/releases/download/v1.0.0/quicktodo-v1.0.0-darwin-arm64.tar.gz"
      sha256 "d77a9d5edceda5f70f3cf6a9a5b658d1202e753ce8f2b57411cb616185c4993b"
    else
      url "https://github.com/justingosan/quicktodo/releases/download/v1.0.0/quicktodo-v1.0.0-darwin-amd64.tar.gz"
      sha256 "39900c468a985bf9bad7e3a1f76ca822e443534e5f71c6b0945cf21a85ad3206"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/justingosan/quicktodo/releases/download/v1.0.0/quicktodo-v1.0.0-linux-arm64.tar.gz"
      sha256 "6b9aef20495b60618e2f138563edfff402fc251af3ecacda4fd8294d4aa7547b"
    else
      url "https://github.com/justingosan/quicktodo/releases/download/v1.0.0/quicktodo-v1.0.0-linux-amd64.tar.gz"
      sha256 "dd86495b8493fb5e8b3d1f35813c68691d70702e3f41444dd05f1479a629661a"
    end
  end

  def install
    if OS.mac?
      if Hardware::CPU.arm?
        bin.install "quicktodo-darwin-arm64" => "quicktodo"
      else
        bin.install "quicktodo-darwin-amd64" => "quicktodo"
      end
    else
      if Hardware::CPU.arm?
        bin.install "quicktodo-linux-arm64" => "quicktodo"
      else
        bin.install "quicktodo-linux-amd64" => "quicktodo"
      end
    end
  end

  test do
    system "#{bin}/quicktodo", "version"
  end
end