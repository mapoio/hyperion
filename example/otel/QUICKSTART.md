# Quick Start Guide

Get up and running with the Hyperion OTel Example in 3 minutes!

## Prerequisites

- Docker (for HyperDX)
- Go 1.24+
- Make

## Option 1: Automated Demo (Recommended)

Run the complete demo with one command:

```bash
./demo.sh
```

This will:
1. ✅ Start HyperDX
2. ✅ Start the example application
3. ✅ Generate sample traffic (successful requests, slow requests, errors)
4. ✅ Open HyperDX UI in your browser

## Option 2: Manual Steps

### Step 1: Start HyperDX

```bash
make hyperdx-up
```

Wait ~30 seconds for HyperDX to fully start.

### Step 2: Run the Application

```bash
make run
```

The server will start on http://localhost:8090

### Step 3: Generate Traffic

In another terminal:

```bash
make load
```

### Step 4: View in HyperDX

Open http://localhost:8080 and explore your traces!

## What to Look For in HyperDX

### 1. Service Map (Coming Soon)
- Your service: `hyperion-otel-example`

### 2. Traces
- Search for recent traces
- Click on a trace to see:
  - Root span: `GET /api/users/:id`
  - Child span: `fetchUser`
  - Child span: `processUser`

### 3. Logs
- Click on a trace
- See correlated logs with same `trace_id`
- Example log entry:
  ```json
  {
    "level": "info",
    "msg": "fetching user",
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
    "span_id": "00f067aa0ba902b7",
    "user_id": "123"
  }
  ```

### 4. Errors
- Search for traces with errors
- See error details in span
- Correlated error logs

## Test Endpoints

```bash
# Health check
curl http://localhost:8090/health

# Get user (creates trace with 2 child spans)
curl http://localhost:8090/api/users/123

# Slow endpoint (2 seconds)
curl http://localhost:8090/api/slow

# Error endpoint (simulates error)
curl http://localhost:8090/api/error
```

## Cleanup

```bash
make clean
```

This stops HyperDX and cleans up Docker volumes.

## Troubleshooting

**Application won't start:**
```bash
# Check if port 8090 is in use
lsof -i :8090

# Kill the process if needed
kill -9 <PID>
```

**HyperDX UI not showing data:**
- Wait 10-30 seconds for data to be ingested
- Ensure the time range filter includes recent data
- Generate more traffic: `make load`

**Docker errors:**
```bash
# Restart Docker daemon
# Then try again:
make clean
make hyperdx-up
```

## Next Steps

See [README.md](README.md) for:
- Detailed architecture explanation
- Configuration options
- Code patterns
- Advanced testing scenarios

## Resources

- [HyperDX Docs](https://www.hyperdx.io/docs)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Hyperion Framework](https://github.com/mapoio/hyperion)
