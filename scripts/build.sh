#!/usr/bin/env bash

if [ "" != "$1" ] 
  then
   rm -rf ./vendor/libgit2/build
fi   

./scripts/build-libgit2-static.sh

./scripts/with-static.sh go build .