#!/usr/bin/env just --justfile


VERSION_FIELD := "Version"
VERSION_FILE := "version.go"

# set version, git commit, git tag
release version:
    @just _echo "== Release binfiled to version {{version}}"
    sed -i \
    's/{{VERSION_FIELD}} = ".*"/{{VERSION_FIELD}} = "{{version}}"/' \
    {{VERSION_FILE}}
    git add {{VERSION_FILE}}
    git commit -m 'bump version to {{version}}'
    git tag {{version}}

_echo msg:
    @echo "\033[0;32;1m{{msg}}\033[0m"