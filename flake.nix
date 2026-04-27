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
            mdformat
            alejandra
            prettier

            # Others
            go-task
          ];
        };
        packages.default = pkgs.buildGoModule {
          pname = "mremotedec";
          version = "0.1.0";
          src = self;
          vendorHash = "sha256-1rWbo42EpOG/xy2+GexjqAdNnONcA9KP0qdwHEI4doo=";
        };
      }
    );
}
