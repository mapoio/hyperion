#!/bin/bash

# test.sh - Automated testing script for the example application

set -e

echo "========================================="
echo "Hyperion OTel Example - Test Script"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if a port is open
check_port() {
    local port=$1
    nc -z localhost $port > /dev/null 2>&1
    return $?
}

# Function to wait for service
wait_for_service() {
    local name=$1
    local port=$2
    local max_wait=60
    local count=0

    echo -n "Waiting for $name to be ready..."
    while ! check_port $port; do
        sleep 1
        count=$((count + 1))
        if [ $count -ge $max_wait ]; then
            echo -e " ${RED}TIMEOUT${NC}"
            return 1
        fi
        echo -n "."
    done
    echo -e " ${GREEN}OK${NC}"
    return 0
}

# Test 1: Check if HyperDX is running
echo "Test 1: Checking HyperDX status..."
if ! docker ps | grep -q hyperdx-local; then
    echo -e "${RED}✗ HyperDX is not running${NC}"
    echo "Run: make hyperdx-up"
    exit 1
fi
echo -e "${GREEN}✓ HyperDX is running${NC}"
echo ""

# Test 2: Check if application is running
echo "Test 2: Checking application status..."
if ! check_port 8090; then
    echo -e "${RED}✗ Application is not running${NC}"
    echo "Run: make run"
    exit 1
fi
echo -e "${GREEN}✓ Application is running${NC}"
echo ""

# Test 3: Health check
echo "Test 3: Health check endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/health)
if [ "$response" == "200" ]; then
    echo -e "${GREEN}✓ Health check passed (HTTP 200)${NC}"
else
    echo -e "${RED}✗ Health check failed (HTTP $response)${NC}"
    exit 1
fi
echo ""

# Test 4: User endpoint
echo "Test 4: Testing /api/users/:id endpoint..."
response=$(curl -s http://localhost:8090/api/users/123)
if echo "$response" | grep -q "John Doe"; then
    echo -e "${GREEN}✓ User endpoint working${NC}"
else
    echo -e "${RED}✗ User endpoint failed${NC}"
    echo "Response: $response"
    exit 1
fi
echo ""

# Test 5: Error endpoint
echo "Test 5: Testing /api/error endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/error)
if [ "$response" == "500" ]; then
    echo -e "${GREEN}✓ Error endpoint working (expected HTTP 500)${NC}"
else
    echo -e "${RED}✗ Error endpoint failed (HTTP $response, expected 500)${NC}"
    exit 1
fi
echo ""

# Test 6: Slow endpoint
echo "Test 6: Testing /api/slow endpoint (this will take 2+ seconds)..."
start_time=$(date +%s)
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/slow)
end_time=$(date +%s)
duration=$((end_time - start_time))

if [ "$response" == "200" ] && [ $duration -ge 2 ]; then
    echo -e "${GREEN}✓ Slow endpoint working (HTTP 200, took ${duration}s)${NC}"
else
    echo -e "${RED}✗ Slow endpoint failed${NC}"
    exit 1
fi
echo ""

# Test 7: Generate load for traces
echo "Test 7: Generating sample traffic for traces..."
for i in {1..5}; do
    curl -s http://localhost:8090/api/users/$i > /dev/null
    echo -n "."
done
echo -e " ${GREEN}OK${NC}"
echo ""

# Test 8: Check HyperDX UI
echo "Test 8: Checking HyperDX UI..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080)
if [ "$response" == "200" ]; then
    echo -e "${GREEN}✓ HyperDX UI is accessible at http://localhost:8080${NC}"
else
    echo -e "${YELLOW}⚠ HyperDX UI returned HTTP $response${NC}"
fi
echo ""

# Summary
echo "========================================="
echo -e "${GREEN}All tests passed!${NC}"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Open HyperDX UI: http://localhost:8080"
echo "2. Search for traces with service name: hyperion-otel-example"
echo "3. Click on a trace to see spans and logs"
echo "4. Verify trace_id appears in logs"
echo ""
echo "To generate more traffic:"
echo "  make load"
echo ""
