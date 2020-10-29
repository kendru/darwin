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
            tags+=("$2")
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

title="$1"
ts=$(date +"%Y%m%d%H%M")
filename="${root_dir}/txt/zet/${ts}-${title}.md"

echo "Creating new note at: $filename"
echo "Using tags: ${tags[@]}"

cat <<EOF > ${filename}
---
tags: ${tags[@]}
created: $(date)
---

# ${title}


EOF

code "$filename "
