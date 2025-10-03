#!/bin/bash

# demo.sh - Complete demo workflow for the Hyperion OTel example

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Banner
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                        ║${NC}"
echo -e "${BLUE}║   Hyperion OTel Example with HyperDX Demo             ║${NC}"
echo -e "${BLUE}║                                                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Step 1: Start HyperDX
echo -e "${YELLOW}Step 1/5: Starting HyperDX...${NC}"
echo ""
make hyperdx-up
echo ""
echo -e "${GREEN}✓ HyperDX started successfully${NC}"
echo -e "  UI: ${BLUE}http://localhost:8080${NC}"
echo ""

# Wait for HyperDX to be fully ready
echo -e "${YELLOW}Waiting for HyperDX to be fully initialized (30 seconds)...${NC}"
for i in {1..30}; do
    echo -n "."
    sleep 1
done
echo ""
echo -e "${GREEN}✓ HyperDX is ready${NC}"
echo ""

# Step 2: Start the application
echo -e "${YELLOW}Step 2/5: Starting the example application...${NC}"
echo ""
echo -e "Starting HTTP server on ${BLUE}http://localhost:8090${NC}"
echo ""

# Start the application in background
go run cmd/app/main.go > app.log 2>&1 &
APP_PID=$!

# Wait for application to start
echo -n "Waiting for application to start"
for i in {1..15}; do
    if curl -s http://localhost:8090/health > /dev/null 2>&1; then
        echo ""
        echo -e "${GREEN}✓ Application started successfully (PID: $APP_PID)${NC}"
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# Step 3: Generate sample traffic
echo -e "${YELLOW}Step 3/5: Generating sample traffic...${NC}"
echo ""

echo "→ Sending 5 successful user requests..."
for i in {1..5}; do
    curl -s http://localhost:8090/api/users/$i > /dev/null
    echo "  ✓ GET /api/users/$i"
    sleep 0.3
done
echo ""

echo "→ Sending 3 slow requests (2 seconds each)..."
for i in {1..3}; do
    curl -s http://localhost:8090/api/slow > /dev/null &
    echo "  ⏱  GET /api/slow (running in background)"
    sleep 0.5
done
echo "  Waiting for slow requests to complete..."
wait
echo -e "${GREEN}  ✓ All slow requests completed${NC}"
echo ""

echo "→ Sending 2 error requests..."
for i in {1..2}; do
    curl -s http://localhost:8090/api/error > /dev/null
    echo "  ✗ GET /api/error (expected error)"
    sleep 0.3
done
echo ""

echo -e "${GREEN}✓ Sample traffic generated successfully${NC}"
echo ""

# Step 4: Show logs with trace context
echo -e "${YELLOW}Step 4/5: Viewing application logs with trace context...${NC}"
echo ""
echo "Recent log entries (showing trace_id and span_id):"
echo "─────────────────────────────────────────────────────"
tail -n 20 app.log | grep -E "trace_id|span_id" || echo "Logs are being buffered..."
echo "─────────────────────────────────────────────────────"
echo ""

# Step 5: Instructions for HyperDX UI
echo -e "${YELLOW}Step 5/5: View data in HyperDX UI${NC}"
echo ""
echo -e "${GREEN}✓ Demo setup complete!${NC}"
echo ""
echo "═══════════════════════════════════════════════════════"
echo -e "${BLUE}Next Steps:${NC}"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "1. Open HyperDX UI:"
echo -e "   ${BLUE}http://localhost:8080${NC}"
echo ""
echo "2. In HyperDX, you should see:"
echo "   • Service: hyperion-otel-example"
echo "   • Multiple traces with different operations"
echo "   • Traces with child spans (fetchUser, processUser)"
echo "   • Error traces (from /api/error endpoint)"
echo "   • Slow traces (from /api/slow endpoint)"
echo ""
echo "3. Click on a trace to:"
echo "   • See the full trace timeline"
echo "   • View span attributes"
echo "   • See correlated logs with same trace_id"
echo ""
echo "4. Generate more traffic:"
echo -e "   ${BLUE}make load${NC}"
echo ""
echo "5. View application logs:"
echo -e "   ${BLUE}tail -f app.log${NC}"
echo ""
echo "6. When done, cleanup:"
echo -e "   ${BLUE}make clean${NC}"
echo ""
echo "═══════════════════════════════════════════════════════"
echo ""

# Open HyperDX UI automatically
echo -n "Opening HyperDX UI in your browser..."
sleep 2
if command -v open > /dev/null; then
    open http://localhost:8080
elif command -v xdg-open > /dev/null; then
    xdg-open http://localhost:8080
fi
echo " Done!"
echo ""

echo -e "${GREEN}Demo is running!${NC}"
echo ""
echo "Application is running with PID: $APP_PID"
echo "To stop the application: kill $APP_PID"
echo "Or use: make clean"
echo ""
