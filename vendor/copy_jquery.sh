#!/bin/bash
set -x -e
if [ -z "$1" ]; then
        echo "Please provide a path to the jquery package root"
        exit 1
fi
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
cd $1
cp -r dist/* ../../../webcontent/static/js/jquery/
