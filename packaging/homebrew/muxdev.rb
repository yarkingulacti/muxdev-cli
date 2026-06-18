class Muxdev < Formula
  desc "Multiplexed dev stack runner"
  homepage "https://github.com/yarkingulacti/muxdev-cli"
  version "VERSION"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yarkingulacti/muxdev-cli/releases/download/vVERSION/muxdev_VERSION_darwin_arm64.tar.gz"
      sha256 "DARWIN_ARM64_SHA256"
    else
      url "https://github.com/yarkingulacti/muxdev-cli/releases/download/vVERSION/muxdev_VERSION_darwin_amd64.tar.gz"
      sha256 "DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yarkingulacti/muxdev-cli/releases/download/vVERSION/muxdev_VERSION_linux_arm64.tar.gz"
      sha256 "LINUX_ARM64_SHA256"
    else
      url "https://github.com/yarkingulacti/muxdev-cli/releases/download/vVERSION/muxdev_VERSION_linux_amd64.tar.gz"
      sha256 "LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "muxdev"
  end

  test do
    assert_match "VERSION", shell_output("#{bin}/muxdev version --short")
  end
end
