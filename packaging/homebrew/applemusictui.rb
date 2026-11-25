class Applemusictui < Formula
  desc "Terminal player for Apple Music"
  homepage "https://github.com/sebasreateguib/AppleMusicTUI"
  url "https://github.com/sebasreateguib/AppleMusicTUI/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"

  depends_on "go" => :build
  depends_on "cava"
  depends_on macos: :sonoma

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
  end

  def caveats
    <<~EOS
      On first use macOS asks to let your terminal control Music and, for the
      visualizer (macOS 14.2 or later), record system audio. Both can be changed in
      System Settings > Privacy & Security.
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/applemusictui --version")
  end
end
