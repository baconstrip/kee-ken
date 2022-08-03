#!/bin/bash

go build
./kiken --static-path="../../../../webcontent/static" --template-path="../../../../webcontent/templates" $@
