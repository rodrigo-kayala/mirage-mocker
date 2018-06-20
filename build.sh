#!/usr/bin/env bash

platforms=("linux/amd64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='mirage-mocker-'$GOOS'-'$GOARCH

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name
done
