
#!/bin/bash

# Load test script for URL Shortener Service
# Requires: vegeta (go install github.com/tsenart/vegeta@latest)

BASE_URL="http://localhost:8080"
RATE="1000"
DURATION="30s"

echo "Starting load test for URL Shortener Service..."
echo "Base URL: $BASE_URL"
echo "Rate: $RATE requests/second"
echo "Duration: $DURATION"

# Create test data
cat > targets.txt << EOF
POST $BASE_URL/api/v1/shorten
Content-Type: application/json

{"url": "https://example.com/test1"}
@@
POST $BASE_URL/api/v1/shorten
Content-Type: application/json

{"url": "https://google.com/test2"}
@@
POST $BASE_URL/api/v1/shorten
Content-Type: application/json

{"url": "https://github.com/test3"}
EOF

# Run shorten endpoint test
echo "Testing URL shortening endpoint..."
vegeta attack -targets=targets.txt -rate=$RATE -duration=$DURATION | vegeta report -type=text

# Test redirect performance (assuming we have some short codes)
echo "Testing redirect performance..."
echo "GET $BASE_URL/abc123" | vegeta attack -rate=$RATE -duration=10s | vegeta report -type=text

# Cleanup
rm -f targets.txt

echo "Load test completed!"
