#!/bin/env bash

##
## Creates a new timestamped note in the Zettelkasten using
## a basic template.
##
## Usage:
##  mkzet some-title
##

root_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

usage() {
    cat <<EOF
Usage:
    $0 some-title
EOF
}

positional=()
tags=()
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -t|--tag)
            tags+=('"'"$2"'"')
            shift
            ;;
        *)
            positional+=("$1")
            ;;
    esac
    shift
done
set -- "${positional[@]}"

if [[ "$#" -ne "1" ]] ; then
    usage
    exit 1
fi

title="$(echo "$1" | perl -pe 's/^/ /; s/[-\h](\w+)/ \u$1/g; s/^ //')"
file_title="$(echo "$1" | perl -pe 's/[^-_a-zA-Z0-9\n]+/-/g; s/([A-Z])/\l$1/g; s/-+$//')"
ts=$(date +"%Y%m%d%H%M")
filename="${root_dir}/txt/zet/${ts}-${file_title}.md"

echo "Creating new note at: $filename"
echo "Using tags: ${tags[@]}"

cat <<EOF > ${filename}
---
tags: [$(IFS=, ; echo "${tags[*]}")]
created: $(date)
---

# ${title}


EOF

if command -v code ; then
	edit=code
elif command -v nvim ; then
	edit=nvim
fi

if [[ -z "$edit" ]] ; then
	edit=${EDITOR:-vim}
fi

$edit "$filename"
