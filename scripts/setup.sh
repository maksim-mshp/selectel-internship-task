set -e

REPO_URL="https://github.com/maksim-mshp/selectel-internship-task"
MODULE_PATH="github.com/maksim-mshp/selectel-internship-task"
PLUGIN_DIR="tools/loglint"

if ! command -v git >/dev/null 2>&1; then
	echo "git not found"
	exit 1
fi

if ! command -v golangci-lint >/dev/null 2>&1; then
	echo "golangci-lint not found"
	exit 1
fi

if [ ! -d "$PLUGIN_DIR" ]; then
	mkdir -p "$PLUGIN_DIR"
	git clone "$REPO_URL" "$PLUGIN_DIR"
fi

gcl_ver=$(golangci-lint version 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="version"){print $(i+1); exit}}')
if [ -z "$gcl_ver" ]; then
	echo "Failed to detect golangci-lint version"
	exit 1
fi

case "$gcl_ver" in
v*) gcl_ver="$gcl_ver" ;;
*) gcl_ver="v$gcl_ver" ;;
esac

cat > .custom-gcl.yml <<EOF
version: $gcl_ver
plugins:
  - module: $MODULE_PATH
    import: $MODULE_PATH/gclplugin
    path: $PLUGIN_DIR
EOF

cat > .golangci.yml <<EOF
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        type: module
        description: log message checks
        config: .loglint.yml
EOF

cat > .loglint.yml <<EOF
lowercase: true
english: true
special: true
sensitive: true
patterns:
  - '(?i)password'
  - '(?i)api_key'
  - '(?i)token'
EOF
