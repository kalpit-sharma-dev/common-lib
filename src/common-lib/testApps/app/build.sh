#!/bin/bash

#BUILDCOMMITSHA=`git rev-parse HEAD`
#TODAY=`date +%Y-%m-%d.%H:%M:%S`
#LDFLAGBUILDVERSION="-X gitlab.kksharmadevdev.com/platform/platform-common-lib/src/app.BuildCommitSHA=$BUILDCOMMITSHA -X github.com/ContinuumLLC/platform-common-lib/src/app.CompiledOn=$TODAY"
#echo $TODAY
#go build -ldflags "$LDFLAGBUILDVERSION"

FILE_PATH=/home/juno/Continuum/workspace/src/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/testApps/app/versioninfo.json
cp $FILE_PATH /home/juno/Continuum/workspace/src/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/app/generate/versioninfo.json
go generate gitlab.kksharmadevdev.com/platform/platform-common-lib/src/app/generate
rm /home/juno/Continuum/workspace/src/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/app/generate/versioninfo.json

