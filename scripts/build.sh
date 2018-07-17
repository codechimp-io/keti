#!/usr/bin/env bash

set -e

repo_path="github.com/codechimp-io/keti"

name="kēti"
version=$( git describe --tags --abbrev=0 2> /dev/null || echo 'unknown' )
revision=$( git rev-parse --short HEAD 2> /dev/null || echo 'unknown' )
branch=$( git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown' )
build_date=$( date -u +%Y-%m-%dT%H:%M:%SZ )
go_version=$( go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/' )

if [ "$(go env GOOS)" = "windows" ]; then
	ext=".exe"
fi

ldflags="
  -s -w
  -X ${repo_path}/version.Name=${name}
  -X ${repo_path}/version.Version=${version}
  -X ${repo_path}/version.Revision=${revision}
  -X ${repo_path}/version.Branch=${branch}
  -X ${repo_path}/version.BuildDate=${build_date}
  -X ${repo_path}/version.GoVersion=${go_version}"

export GO15VENDOREXPERIMENT="1"

echo " >   keti"
CGO_ENABLED=0 go build -a -ldflags "${ldflags}" -o keti${ext} ${repo_path}/cmd/keti

exit 0
