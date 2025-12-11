# typed: false
# frozen_string_literal: true

class Versionator < Formula
  desc "A semantic version management CLI tool"
  homepage "https://github.com/benjaminabbitt/versionator"
  version "$VERSION$"
  license "BSD-3-Clause"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/benjaminabbitt/versionator/releases/download/v$VERSION$/versionator-darwin-arm64"
      sha256 "$SHA256_DARWIN_ARM64$"

      def install
        bin.install "versionator-darwin-arm64" => "versionator"
      end
    else
      url "https://github.com/benjaminabbitt/versionator/releases/download/v$VERSION$/versionator-darwin-amd64"
      sha256 "$SHA256_DARWIN_AMD64$"

      def install
        bin.install "versionator-darwin-amd64" => "versionator"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/benjaminabbitt/versionator/releases/download/v$VERSION$/versionator-linux-arm64"
      sha256 "$SHA256_LINUX_ARM64$"

      def install
        bin.install "versionator-linux-arm64" => "versionator"
      end
    else
      url "https://github.com/benjaminabbitt/versionator/releases/download/v$VERSION$/versionator-linux-amd64"
      sha256 "$SHA256_LINUX_AMD64$"

      def install
        bin.install "versionator-linux-amd64" => "versionator"
      end
    end
  end

  test do
    # Create a test VERSION file
    (testpath/"VERSION").write("1.0.0")
    assert_match "1.0.0", shell_output("#{bin}/versionator version")
  end
end
