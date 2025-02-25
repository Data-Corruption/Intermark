MOD_NAME="intermark"
VERSION_VAR_PATH="$MOD_NAME/internal/app.Version"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENTRY_POINT="$PROJECT_ROOT/cmd/$MOD_NAME"
BIN_DIR="$PROJECT_ROOT/bin"

# Stop on error
set -e
set -o pipefail

# Clean binary directory
if [ -d "$BIN_DIR" ]; then
  rm -rf "$BIN_DIR"
fi
mkdir -p "$BIN_DIR"
echo "Cleaned binary directory."

# Set the version if it's not set
if [ -z "$VERSION" ]; then
  VERSION="vX.X.X"
fi
echo "Output version set to $VERSION"

# Function to build a binary
build() {
  export GOOS=$1; export GOARCH=$2
  OUTPUT_PATH="$BIN_DIR/$MOD_NAME-$1-$2"
  if [ "$1" == "windows" ]; then
    OUTPUT_PATH="$OUTPUT_PATH.exe"
  fi
  eval "go build -ldflags=\"-X '$VERSION_VAR_PATH=$VERSION'\" -o \"$OUTPUT_PATH\" \"$ENTRY_POINT\""
  echo "Built $MOD_NAME for $1 $2."
}

# For a list of supported GOOS and GOARCH values, see https://golang.org/doc/install/source#environment
build linux amd64
build windows amd64