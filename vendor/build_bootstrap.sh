#!/bin/bash
set -x 
if [ -z "$1" ]; then
	echo "Please provide a path to the bootstrap files"
	exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
cd $1
sudo npm install
npm run dist

mkdir -p ../../webcontent/static/css/vendor/bootstrap
mkdir -p ../../webcontent/static/js/vendor/bootstrap

cp -r dist/css/* ../../webcontent/static/css/vendor/bootstrap
cp -r dist/js/* ../../webcontent/static/js/vendor/bootstrap
