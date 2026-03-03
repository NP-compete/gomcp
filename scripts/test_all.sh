#!/bin/bash

# MCP 2025-06-18 Comprehensive Test Suite
# Tests all features: completion, logging, pagination, tools, prompts, resources, roots, etc.

set -e

echo "MCP 2025-06-18 Test Suite"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Change to gomcp directory
cd "$(dirname "$0")/.."

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run tests for a package
run_test() {
    local package=$1
    local description=$2
    
    echo -e "${YELLOW}Testing:${NC} $description"
    
    if go test -v "./$package" 2>&1 | tee /tmp/test_output.log; then
        echo -e "${GREEN}PASSED${NC}: $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}: $description"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        cat /tmp/test_output.log
    fi
    echo ""
}

echo "Unit Tests"
echo "-------------"
echo ""

# Completion tests
run_test "internal/completion" "Completion (Structured Outputs)"

# Logging tests
run_test "internal/logging" "Logging (Server to Client Notifications)"

# Pagination tests
run_test "internal/pagination" "Pagination (Cursor-Based)"

# Roots tests
run_test "internal/roots" "Roots (Filesystem Boundaries)"

echo ""
echo "Integration Tests"
echo "-------------------"
echo ""

# Integration tests
run_test "test" "Full MCP Protocol Integration"

echo ""
echo "Test Summary"
echo "==============="
echo ""

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))

echo "Total tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

# Coverage report
echo "Generating Coverage Report"
echo "=============================="
echo ""

go test -coverprofile=coverage.out ./internal/completion ./internal/logging ./internal/pagination ./internal/roots ./test
go tool cover -html=coverage.out -o coverage.html

echo -e "${GREEN}Coverage report generated:${NC} coverage.html"
echo ""

# Exit with error if any tests failed
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
