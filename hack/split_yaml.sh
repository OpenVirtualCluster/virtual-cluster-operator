#!/bin/bash

input_file="chart/templates/manager.yaml"
output_dir=$(dirname "$input_file")
filename=$(basename "$input_file" .yaml)

split_yaml() {
    awk -v output_dir="$output_dir" -v filename="$filename" '
    BEGIN {
        file_index = 0
        kind = "unknown"
        name = "unknown"
        file_content = ""
    }
    function write_file() {
        if (file_content != "") {
            kind = tolower(kind)
            output_file = output_dir "/" filename "_" kind "_" name ".yaml"
            print file_content > output_file
        }
    }
    /^---$/ {
        write_file()
        file_index++
        kind = "unknown"
        name = "unknown"
        file_content = ""
        next
    }
    /^[^[:space:]]+:/ {
        if ($1 == "kind:") {
            kind = $2
            gsub(/[ \t]+/, "", kind)
        }
    }
    /metadata:/ {
        in_metadata = 1
    }
    in_metadata && /^[[:space:]]+name:/ {
        split($0, arr, ":")
        name = arr[2]
        gsub(/[ \t]+/, "", name)
        in_metadata = 0
    }
    {
        file_content = file_content $0 "\n"
    }
    END {
        write_file()
    }
    ' "$input_file"
}

split_yaml
echo "Files have been split and saved in the directory: $output_dir"