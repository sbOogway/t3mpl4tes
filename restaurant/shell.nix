{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go           
  ];

  shellHook = ''
    export GOPATH=~/.go
    export PATH=$GOPATH/bin:$PATH
    mkdir -p $GOPATH

    if [ ! -f go.mod ]; then
      go mod init myproject  # Replace 'myproject' with your module name
    fi

    # Add Gin as a dependency
    go get -u github.com/gin-gonic/gin
    go get -u github.com/k0kubun/pp/v3
    go get go.mongodb.org/mongo-driver/mongo
    go get go.mongodb.org/mongo-driver/mongo/options
    go get go.mongodb.org/mongo-driver/bson

    export GIN_MODE=release

    go run main.go
  '';
}
