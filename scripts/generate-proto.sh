#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/backend-crud-proto.XXXXXX")"
GO_BIN_DIR="${TMP_DIR}/bin"
PYTHON_TOOLS_DIR="${TMP_DIR}/python-tools"
PROTO_FILE="proto/embedding/v1/embedding.proto"

PROTOC_GEN_GO_VERSION="v1.36.5"
PROTOC_GEN_GO_GRPC_VERSION="v1.5.1"
GRPCIO_TOOLS_VERSION="1.71.0"

cleanup() {
	rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

require_command go
require_command python3

mkdir -p "${GO_BIN_DIR}"

echo "Installing pinned protobuf generators..."
GOBIN="${GO_BIN_DIR}" go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
GOBIN="${GO_BIN_DIR}" go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"
python3 -m pip install --quiet --target "${PYTHON_TOOLS_DIR}" "grpcio-tools==${GRPCIO_TOOLS_VERSION}"

echo "Generating Go protobuf stubs..."
PYTHONPATH="${PYTHON_TOOLS_DIR}" python3 -m grpc_tools.protoc \
	-I"${ROOT_DIR}" \
	--plugin=protoc-gen-go="${GO_BIN_DIR}/protoc-gen-go" \
	--plugin=protoc-gen-go-grpc="${GO_BIN_DIR}/protoc-gen-go-grpc" \
	--go_out="${ROOT_DIR}" \
	--go_opt=module=github.com/IanStuardo-Dev/backend-crud \
	--go-grpc_out="${ROOT_DIR}" \
	--go-grpc_opt=module=github.com/IanStuardo-Dev/backend-crud \
	"${PROTO_FILE}"

echo "Generating Python protobuf stubs..."
PYTHONPATH="${PYTHON_TOOLS_DIR}" python3 -m grpc_tools.protoc \
	-I"${ROOT_DIR}" \
	--python_out="${ROOT_DIR}/services/embedding-service" \
	--grpc_python_out="${ROOT_DIR}/services/embedding-service" \
	"${PROTO_FILE}"

touch \
	"${ROOT_DIR}/services/embedding-service/proto/__init__.py" \
	"${ROOT_DIR}/services/embedding-service/proto/embedding/__init__.py" \
	"${ROOT_DIR}/services/embedding-service/proto/embedding/v1/__init__.py"

echo "Protobuf stubs regenerated."
