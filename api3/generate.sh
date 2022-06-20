#!/bin/bash	


(cd ../tools/spec-aggregator && go run main.go)
(cd ./common && go generate)
(cd ./global && go generate)
(cd ./notifications && go generate)

if [[ $1 == "--verify" ]]  ; then
    diffs=$(git status --porcelain)
    if [[ ${diffs} != "" ]]; then
        echo "Verification Failed: this script was not run on the latest version of the specs"
        git diff
        exit 1
    fi
fi