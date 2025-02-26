#!/bin/bash

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Extract coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | tr -d '%')

# Create badge JSON file
cat > badge.json << EOF
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "${COVERAGE}%",
  "color": "brightgreen"
}
EOF

echo "Coverage: ${COVERAGE}%"
echo "Coverage badge data generated in badge.json"
echo "Coverage HTML report generated in coverage.html"

# Update README badge (optional)
# This will replace the coverage percentage in the README.md file
if [ -f "README.md" ]; then
  sed -i "s/Coverage-[0-9.]*%25/Coverage-${COVERAGE}%25/g" README.md
  echo "README.md coverage badge updated"
fi