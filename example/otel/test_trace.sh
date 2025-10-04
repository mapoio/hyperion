#!/bin/bash

# Start the application
go run cmd/app/main.go > /tmp/app.log 2>&1 &
APP_PID=$!

# Wait for server to start
sleep 3

# Make first request
echo "=== Making first request ==="
curl -s -X POST http://localhost:8090/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123","product_id":"prod456","amount":99.99}'
echo ""

sleep 2

# Make second request
echo "=== Making second request ==="
curl -s -X POST http://localhost:8090/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user456","product_id":"prod789","amount":199.99}'
echo ""

sleep 2

# Kill the application
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo "=== Checking trace context in logs ==="
grep -E "(trace_id|span_id|creating order)" /tmp/app.log | head -20
