#!/bin/bash

cd ./../go-ast-matcher
./update.sh
git commit -am "update go.mod" && git push origin
cd -

cd ./../go-callgraph
./update.sh
git commit -am "update go.mod" && git push origin
cd -

cd ./../go-imports
./update.sh
git commit -am "update go.mod" && git push origin
cd -