{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {

  shellHook = ''
    echo "+-----------------------------------------------+"
    echo "| YOU ARE NOW IN THE APICLARITY DEV ENVIRONMENT |"
    echo "+-----------------------------------------------+"
  '';

  buildInputs = [
    pkgs.nodejs-17_x
    pkgs.go_1_17
    pkgs.go-swagger
    pkgs.kubectl
    pkgs.kubernetes-helm
  ];
}
