{
  description = "AppleMusicTUI — Terminal UI for Apple Music";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-darwin" "aarch64-darwin" "x86_64-linux" ];
      eachSystem = f: nixpkgs.lib.genAttrs systems (system:
        f (import nixpkgs { inherit system; }));
    in
    {
      packages = eachSystem (pkgs: rec {
        default = music-player;
        music-player = pkgs.buildGoModule {
          pname = "music-player";
          version = "unstable";
          src = ./.;
          vendorHash = "sha256-SMhllO87YlmySHroKfPq1pHb67CwHaZ3XMp3t983etc=";
          # Runtime needs assets (logo) — copy the whole source
          postInstall = ''
            mkdir -p $out/share/music-player
            cp -r assets $out/share/music-player/
          '';
          meta = with pkgs.lib; {
            description = "Terminal UI for Apple Music (macOS)";
            mainProgram = "music-player";
            platforms = platforms.darwin;
          };
        };
      });

      devShells = eachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [ go ];
        };
      });
    };
}
