# README

## Overview

Kee-ken is an open source real-time trivia game.

## License

The project is licensed under the Apache License 2.0. For the full license, 
please see the "LICENSE" file.

## Building and Running

### Dependencies

This program relies on the following packages being installed: 
`npm`
`golang-go`

On Debian derivatives, you can install the required dependencies with:

```sh
apt install golang-go npm
```

This is built and tested with Go version 1.10, earlier versions may not 
compile.

### Building

First clone the project and navigate to the directory you checked it out into.

```sh
git clone https://github.com/baconstrip/kiken
cd ./kiken
```

Then navigate to the `vendor` directory and run the NPM install command. This 
will collect and download JS dependencies.

```sh
cd vendor
npm install
``` 

Once these are installed, you can run the various scripts to build and install 
the JS libraries to the location the server expects them:

```sh
./build_bootstrap.sh bootstrap
./copy_js_deps.sh
```

After doing this, navigate up a directory and enter the server directory, then
set the GOPATH environment variable to this directory:

```sh
cd ../server
export GOPATH=$(pwd)
```

Once this is set, you can navigate to the directory that contains the server
code, install the Go depdenencies, and then start the server with the included
script:

```sh
cd src/github.com/baconstrip/kiken/
go get
./test_in_place.sh --question-list="../../../../../example/example_questions.json"
```

Note the server requires the extra flag "question-list," which instructs the
server where to look for questions. In this case, the example included in this
repository.

