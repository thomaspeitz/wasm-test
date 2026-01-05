#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WASM_FILE="market-header.wasm"
OUTPUT_DIR="${SCRIPT_DIR}/generated"
CONFIGMAP_FILE="${OUTPUT_DIR}/wasm-configmap.yaml"
NAMESPACE="envoy-gateway-system"

echo "🔧 Building WASM plugin with TinyGo..."

# Check for TinyGo
if ! command -v tinygo &> /dev/null; then
    echo "❌ TinyGo not found."
    echo ""
    echo "Install TinyGo:"
    echo "  macOS:  brew install tinygo"
    echo "  Linux:  https://tinygo.org/getting-started/install/"
    echo ""
    exit 1
fi

cd "$SCRIPT_DIR"

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod tidy

# Build the WASM module with TinyGo
echo "🔨 Compiling with TinyGo..."
tinygo build -o "${WASM_FILE}" -scheduler=none -target=wasi ./main.go

# Get file size
WASM_SIZE=$(ls -lh "${WASM_FILE}" | awk '{print $5}')
echo "✅ Built ${WASM_FILE} (${WASM_SIZE})"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Generate ConfigMap with binary data (wrapped for helmfile raw chart)
echo "📦 Generating ConfigMap..."
cat > "${CONFIGMAP_FILE}" << EOF
# Auto-generated - do not edit manually
# Run 'wasm/build.sh' to regenerate
resources:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: wasm-plugins
      namespace: ${NAMESPACE}
    binaryData:
      market-header.wasm: $(base64 < "${WASM_FILE}" | tr -d '\n')
EOF

echo "✅ Generated ${CONFIGMAP_FILE}"

# Show size info
echo ""
echo "📊 Summary:"
echo "   WASM binary:  ${WASM_SIZE}"
echo "   ConfigMap:    $(ls -lh "${CONFIGMAP_FILE}" | awk '{print $5}')"
echo ""
echo "🚀 To deploy:"
echo "   kubectl apply -f ${CONFIGMAP_FILE}"
echo "   helmfile apply"
echo ""
echo "🔍 To verify WASM is loaded in Envoy:"
echo "   kubectl logs -n ${NAMESPACE} -l gateway.envoyproxy.io/owning-gateway-name=default -c envoy | grep -i wasm"
