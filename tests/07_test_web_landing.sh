#!/bin/bash
set -e
echo "-> Test: Web Landing Page"
curl -s http://localhost:8080/ > /tmp/web_out
grep "CromIA" /tmp/web_out > /dev/null
echo "   Passou!"
