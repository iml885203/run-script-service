#!/bin/bash

# Main execution script - runs multiple tasks in sequence
echo "$(date): Starting main execution cycle..."

# Task 1: Run 591 TypeScript migration
echo "$(date): Executing 591 TypeScript migration..."
if [ -f "./run-591-migrate-ts.sh" ]; then
    ./run-591-migrate-ts.sh
    if [ $? -eq 0 ]; then
        echo "$(date): 591 TypeScript migration completed successfully"
    else
        echo "$(date): 591 TypeScript migration failed"
    fi
else
    echo "$(date): run-591-migrate-ts.sh not found, skipping..."
fi

# Task 2: Run plan development cycle
echo "$(date): Executing plan development cycle..."
if [ -f "./run-plan-cycle.sh" ]; then
    ./run-plan-cycle.sh
    if [ $? -eq 0 ]; then
        echo "$(date): Plan development cycle completed successfully"
    else
        echo "$(date): Plan development cycle failed"
    fi
else
    echo "$(date): run-plan-cycle.sh not found, skipping..."
fi

echo "$(date): Main execution cycle completed"