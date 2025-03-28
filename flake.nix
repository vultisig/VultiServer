{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    fenix.url = "github:nix-community/fenix";
    fenix.inputs = { nixpkgs.follows = "nixpkgs"; };
    devenv.url = "github:cachix/devenv";
  };

  nixConfig = {
    extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs = { self, nixpkgs, devenv, systems, ... } @ inputs:
    let
      forEachSystem = nixpkgs.lib.genAttrs (import systems);
    in
    {
      packages = forEachSystem (system: {
        devenv-up = self.devShells.${system}.default.config.procfileScript;
        devenv-test = self.devShells.${system}.default.config.test;
      });

      devShells = forEachSystem
        (system:
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in
          {
            default = devenv.lib.mkShell {
              inherit inputs pkgs;
              modules = [
                {
                  # https://devenv.sh/reference/options/
                  services = {
                    redis = {
                      enable = true;
                    };

                    minio = {
                      enable = true;
                    };

                    postgres = {
                      enable = true;
                      package = pkgs.postgresql_16;
                      initialDatabases = [
                        { name = "vs-plugins-server"; }
                        { name = "vs-plugins-plugin"; }
                      ];
                      listen_addresses = "127.0.0.1";
                      settings = {
                        unix_socket_directories = "/tmp";
                      };
                    };
                  };

                  languages.go = {
                    enable = true;
                  };

                  packages = with pkgs; [ 
                    redis
                    goose
                  ];

                  enterShell = ''
                    echo "vultisig-plugins shell"
                  '';
                }
              ];
            };
          });
    };
}
