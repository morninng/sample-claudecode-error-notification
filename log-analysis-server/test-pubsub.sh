#!/bin/bash

# Test script for sending a Pub/Sub push message to local log-analysis-server



# Sample log entry
LOG_ENTRY=$(cat <<EOF
{
  "severity": "ERROR",
  "textPayload": "error occurred with invalid query message",
  "timestamp": "2025-01-01T12:00:00Z",
  "resource": {
    "type": "cloud_run_revision",
    "labels": {
      "service_name": "api-server",
      "revision_name": "api-server-00001"
    }
  },
  "labels": {}
}
EOF
)

# Encode to base64
BASE64_DATA=$(echo "$LOG_ENTRY" | base64)

# Create Pub/Sub message
PUBSUB_MESSAGE=$(cat <<EOF
{
  "message": {
    "data": "$BASE64_DATA",
    "attributes": {}
  }
}
EOF
)

echo "Sending Pub/Sub push message to http://localhost:8080/"
echo ""

# Send the request
RESPONSE=$(curl -s -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d "$PUBSUB_MESSAGE" \
  -w "\n%{http_code}")

# Extract status code (last line)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

echo "Response:"
echo "HTTP Status: $HTTP_CODE"
echo "Body: $BODY"
echo ""
