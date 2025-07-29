#!/bin/bash
cd /home/logan/591-crawler && claude --dangerously-skip-permissions -p "fix test and push" --output-format stream-json --verbose
echo "$(date): Fix test and push script executed successfully"