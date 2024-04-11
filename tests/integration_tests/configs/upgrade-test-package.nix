let
  pkgs = import ../../../nix { };
  fetchEthermint = rev: builtins.fetchTarball "https://github.com/Kava-Labs/ethermint/archive/${rev}.tar.gz";
  released = pkgs.buildGo121Module rec {
    name = "ethermintd";
    src = fetchEthermint "60c4f850ac0bddce2d584feed4f5ac82c9df7c9c";
    subPackages = [ "cmd/ethermintd" ];
    vendorSha256 = "sha256-5JXFuwzTYPc71uPK8OMtc6dmlF6/Lj4TcMCP3GTG76U=";
    doCheck = false;
  };
  current = pkgs.callPackage ../../../. { };
in
pkgs.linkFarm "upgrade-test-package" [
  { name = "genesis"; path = released; }
  { name = "integration-test-upgrade"; path = current; }
]
