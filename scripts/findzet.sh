#!/bin/env bash

##
## Finds a Zettelkasten note by its title, giving the ID.
##
## This should be replaced by a more flexible search.
##
## Usage:
##  findzet some-title
##

root_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

usage() {
    cat <<EOF
Usage:
    $0 some-title
EOF
}

if [[ "$#" -ne "1" ]] ; then
    usage
    exit 1
fi

ls -l "${root_dir}/txt/zet/" \
    | grep "$1" \
    | awk '{ print $9; }' \
    | awk -F- '{ print $1; }'
