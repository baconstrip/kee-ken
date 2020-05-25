#!/bin/bash
set -x -e
if [ -z "$1" ]; then
	echo "Please provide a path to the bootstrap files"
	exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
cd $1
sudo npm install
npm run dist

cp -r dist/css/* ../../webcontent/static/css/bootstrap
cp -r dist/js/* ../../webcontent/static/js/bootstrap
