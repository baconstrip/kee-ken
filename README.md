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

### Compatibility Information

This is built and tested with Go version 1.18, and requires at least 1.16 to 
compile and run.

It is recommended that you use an older version of node when building this
project, such as version 14.10.0. If using nvm, run:

```
nvm install 14.10.0
nvm use 14.10.0
```

### Building

First clone the project and navigate to the directory you checked it out into.

**NOTE: The project must be cloned using the --recurse-submodules flag!**

```sh
git clone --recurse-submodules https://github.com/baconstrip/kee-ken
cd ./kee-ken
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

Once this is complete, you can navigate to the directory that contains the 
server code, install the Go depdenencies, and then start the server with the 
included script:

```sh
cd server
go get
go build
./test_in_place.sh --question-list="../example/example_questions.json"
```

Note the server requires the extra flag "question-list," which instructs the
server where to look for questions. In this case, the example included in this
repository.

