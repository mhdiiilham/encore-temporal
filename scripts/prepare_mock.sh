#!/usr/bin/env bash
set -euo pipefail

if ! command -v mockgen &> /dev/null; then
    echo "❌ mockgen not found."
    echo "👉 Please install with: go install go.uber.org/mock/mockgen@latest"
    exit 1
fi

echo "mockgen found, generating mocks..."
mockgen -source=billing/usecases/interfaces.go -destination=billing/usecases/mock/interfaces_mock.go
mockgen -source=billing/domain/repository.go -destination=billing/domain/mock/repository_mock.go
mockgen -source=pkg/generator/generator.go -destination=pkg/generator/mock/mock_generator.go
mockgen -source=pkg/clock/clock.go -destination=pkg/clock/mock/mock_clock.go

echo "Mocks generated successfully!"
