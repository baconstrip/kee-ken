#!/bin/bash
set -x -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

mkdir -p ../webcontent/static/js/vendor/jquery/
mkdir -p ../webcontent/static/js/vendor/vue/

cp -r node_modules/jquery/dist/* ../webcontent/static/js/vendor/jquery/
cp -r node_modules/vue/dist/* ../webcontent/static/js/vendor/vue/
