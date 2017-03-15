#!/usr/bin/env bash

if [ "" != "$1" ] 
  then
   rm -rf ./cdeps/libgit2/build
fi   

./scripts/build-libgit2-static.sh

./scripts/with-static.sh go build .