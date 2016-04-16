GODIR=.go
GOPATH=/usr/src/golang
PACKFILE=gopkg
if [ -f $PACKFILE ]; then
    OUT=`cat ${PACKFILE}`
else
    OUT=main
fi
GODEPS=godeps
GOVER=gover

function __init {
    # Project name
    if [ $# -lt 1 ]
    then
        echo -n "Enter project name: "
        read NAME
    else
        NAME=$1
    fi

    # Update FS layout
    if [ ! -f $PACKFILE ]; then
        echo $NAME > $PACKFILE;
    fi

    if [ ! -f "$GODEPS" ]; then
        touch $GODEPS
    fi

    if [ ! -d "$GODIR" ]; then
        mkdir "$GODIR"
    fi

    # Update .gitignore file
    if [ ! -f ".gitignore" ]; then
        touch .gitignore
    fi

    if [ -z "$(cat '.gitignore' | grep "$GODIR")" ]; then
        echo $GODIR >> .gitignore
    fi

    if [ -z "$(cat '.gitignore' | grep "build")" ]; then
        echo "build" >> .gitignore
    fi

    if [ ! -f "$NAME.go" ]; then
      echo 'package main' > "$NAME.go"
    fi

    if [ ! -d ".git" ]; then
      git init
    fi
};

# Install package from repository or file system
# Note that installation is running inside docker container
# and you should to add directory manually
function __install {
    if [ $# -gt 0 ]
    then
        docker run --rm \
    		-e GOPATH=${GOPATH} \
    		-v $PWD/$GODIR:${GOPATH} \
    		-w $GOPATH/ \
    		golang go get "$1" || exit 1

        if [ -z "$(cat $GODEPS | grep "$1")" ]; then
            echo $1 >> $GODEPS
        fi
    else
        while read dep
        do
            bake install $dep
        done < $GODEPS
    fi
}

# Remove package from list of dependencies
function __uninstall {
    SRC=$GODIR/src/$1

    if [ -d "$SRC" ]; then
        rm -rf $SRC
        rm -rf $GODIR/pkg/*/$1.a

        cat $GODEPS | grep -v "$1" > $GODEPS
    fi
}

# Build binary from sources.
# The first argument is a name of binary.
function build {
  OUT=$1
  shift 1

  if [ -z "$GOOS" ]
  then
      GOOS=linux
  fi

  if [ -z "$GOARCH" ]
  then
      GOARCH=amd64
  fi

  docker run --rm \
  -e GOPATH=${GOPATH}:/ \
  -v $PWD/$GODIR:$GOPATH \
  -v $PWD:/src \
  -e GOOS=$GOOS \
  -e GOARCH=$GOARCH \
  -w /src \
  golang go build -o build/$OUT $@
}

# Build default target
function __build {
    build $OUT $@
}

# Build and run
function __run {
    ./build/$OUT $@
}

function __test {
  ./build/$OUT 1999 /tmp/bolt.db &
  PID=$!

  sleep 3

  curl \
    -X POST  \
    -H "Content-Type: application/json" \
    -d '{"id": 1, "method": "Users.Add", "params": [{"name":"user","role":"user", "keys":[{"type":"ed25519", "name":"ssh", "value":"12345678901234567890123456789012"}]}]}' \
    http://localhost:1999/rpc

  curl \
    -X POST  \
    -H "Content-Type: application/json" \
    -d '{"username": "user", "signature": "1234567890", "type":"authenticate"}' \
    http://localhost:1999/

  kill -s 9 $PID
}

function __release {
  __clean

  __release_linux
  __release_darwin
}

function __release_linux {
    __build $@
    tar -cjf build/$OUT-linux_x64.tar.gz -C build $OUT
    rm build/$OUT
}

function __release_darwin {
    GOOS=darwin GOARCH=amd64 __build $@
    tar -cjf build/$OUT-darwin_x64.tar.gz -C build $OUT
    rm build/$OUT
}

# Clean build directory
function __clean {
    rm -rf build/*
}
