{
  description = "AppleMusicTUI - Terminal UI for Apple Music";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-darwin" "aarch64-darwin" ];
      eachSystem = f: nixpkgs.lib.genAttrs systems (system:
        f (import nixpkgs { inherit system; }));
    in {
      packages = eachSystem (pkgs: rec {
        default = music-player;
        music-player = pkgs.buildGoModule {
          pname = "music-player";
          version = "unstable";
          src = ./.;
          vendorHash = "sha256-SMhllO87YlmySHroKfPq1pHb67CwHaZ3XMp3t983etc=";
          postInstall = ''
            mkdir -p $out/share/music-player
            cp -r assets $out/share/music-player/
          '';
          meta = with pkgs.lib; {
            description = "Terminal UI for Apple Music (macOS)";
            mainProgram = "SR-Player";
            platforms = platforms.darwin;
          };
        };
      });
    };
}
