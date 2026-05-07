{
  inputs.utils.url = "github:numtide/flake-utils";

  outputs = {
    self,
    nixpkgs,
    utils,
  }:
    utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShells.default = pkgs.mkShell rec {
          buildInputs = with pkgs; [
            # Go
            go
            gopls
            delve

            # Formatters
            treefmt
            taplo
            alejandra
            prettier

            # Others
            go-task
          ];
        };
        packages.default = pkgs.buildGoModule {
          pname = "mremotedec";
          version = "0.2.0";
          src = self;
          vendorHash = "sha256-dA9Y9cVoGS5nD0fgKbufQ7EZeLs2yxIs3MT9iXnU0K4=";
        };
      }
    );
}
