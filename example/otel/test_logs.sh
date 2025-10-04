#!/bin/bash

# Kill any existing process
pkill -9 -f "go run cmd/app/main.go"
sleep 1

# Start the app in background
go run cmd/app/main.go > /tmp/otel_test.log 2>&1 &
APP_PID=$!

# Wait for server to start
sleep 5

# Make a request
echo "Making request..."
curl -s -X POST http://localhost:8090/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test123","product_id":"prod456","amount":123.45}'

# Wait for logs
sleep 2

# Stop the app
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

# Show logs with trace_id
echo ""
echo "=== Logs with trace_id ==="
cat /tmp/otel_test.log | jq -c 'select(.trace_id != null)' 2>/dev/null | head -10

if [ $? -ne 0 ]; then
    echo "=== Raw JSON logs ==="
    cat /tmp/otel_test.log | grep "creating order" | head -5
fi
