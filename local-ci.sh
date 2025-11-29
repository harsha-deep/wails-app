#!/usr/bin/env bash
#
# Local CI Runner for Wails App (auto-installs golangci-lint if missing)
#
set -euo pipefail

echo "========================================"
echo "🔧 Local CI Started"
echo "========================================"

###########################################
# 1. Build frontend
###########################################
echo "📦 Step 1/9: Installing frontend dependencies..."
cd frontend
npm ci

echo "🏗️  Step 2/9: Building frontend (producing dist/)..."
npm run build
cd ..

###########################################
# 2. Go dependencies
###########################################
echo "📚 Step 3/9: Downloading Go modules..."
go mod download

###########################################
# 3. Run Go tests
###########################################
echo "🧪 Step 4/9: Running Go tests..."
go test -v ./...

###########################################
# 4. gofmt
###########################################
echo "📝 Step 5/9: Checking Go formatting..."
UNFORMATTED=$(gofmt -s -l .)

if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files are not formatted:"
    echo "$UNFORMATTED"
    exit 1
else
    echo "✔️  Formatting OK"
fi

###########################################
# 5. go vet
###########################################
echo "🔍 Step 6/9: Running go vet..."
go vet ./...

###########################################
# 6. Ensure golangci-lint exists (auto-install)
###########################################
echo "🧹 Step 7/9: Ensuring golangci-lint is installed..."

if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found — installing v1.64.0 to $(go env GOPATH)/bin ..."
    # Install to GOPATH/bin (works when GOPATH is default); you can change to GOBIN if you prefer
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.64.0

    # Ensure GOPATH/bin is on PATH for this script run
    export PATH="$(go env GOPATH)/bin:$PATH"

    if ! command -v golangci-lint &> /dev/null; then
        echo "Failed to install golangci-lint or PATH not updated. Please install manually and retry."
        exit 1
    fi
else
    echo "golangci-lint already installed: $(command -v golangci-lint)"
fi

###########################################
# 7. Run golangci-lint
###########################################
echo "🔎 Step 8/9: Running golangci-lint..."
golangci-lint run -c /home/harsha/per/app/.github/workflows/.golangci.yml --timeout=5m

###########################################
# 8. Optional: Run Wails build
###########################################
echo "📦 Step 9/9: (Optional) Wails build..."
if command -v wails &> /dev/null; then
    echo "Running: wails build"
    wails build
else
    echo "Skipping Wails build — Wails CLI not installed."
fi

echo "========================================"
echo "🎉 Local CI Completed Successfully!"
echo "========================================"