class Eventhook < Formula
  desc "Webhook infrastructure runtime — Stripe-quality observability for every event"
  homepage "https://eventhook.dev"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/eventhook/eventhook/releases/download/v0.1.0/eventhook_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/eventhook/eventhook/releases/download/v0.1.0/eventhook_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/eventhook/eventhook/releases/download/v0.1.0/eventhook_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/eventhook/eventhook/releases/download/v0.1.0/eventhook_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "eventhook"
  end

  test do
    assert_match "EventHook", shell_output("#{bin}/eventhook --help")
  end
end
