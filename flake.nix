{
    description = "Dev environment for Coin";

    inputs = {
        flake-utils.url = "github:numtide/flake-utils";

        # 24.5.0
        nodejs-nixpkgs.url = "github:NixOS/nixpkgs/281aac132f6cd84252a5a242cde14c183f600cbc";
    };

    outputs = {
        self,
        flake-utils,
        nodejs-nixpkgs
    } @inputs:
        flake-utils.lib.eachDefaultSystem (system: let
            nodejspkg = nodejs-nixpkgs.legacyPackages.${system};
        in {
            devShells.default = nodejspkg.mkShell {
                packages = [
                    nodejspkg.nodejs_24
                ];

                shellHook = ''
                    echo "node" "$(node --version)"
                    echo "npm" "$(npm --version)"
                    echo "npx" "$(npx --version)"
                '';
            };
        });
}

