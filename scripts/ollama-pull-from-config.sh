#!/bin/sh
set -eu

config_path="${CONFIG_PATH:-/app/config.json}"

model="$(
	awk '
		/"ollama"[[:space:]]*:/ { in_ollama = 1 }
		in_ollama && /"model"[[:space:]]*:/ {
			line = $0
			sub(/^.*"model"[[:space:]]*:[[:space:]]*"/, "", line)
			sub(/".*$/, "", line)
			print line
			exit
		}
	' "$config_path"
)"

if [ -z "$model" ]; then
	echo "ollama.model is required in $config_path" >&2
	exit 1
fi

ollama pull "$model"
