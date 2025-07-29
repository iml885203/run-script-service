#!/usr/bin/env python3
import subprocess
import time
import signal
import sys
import os
import json
from pathlib import Path
from datetime import datetime

class RunScriptService:
    def __init__(self):
        self.running = False
        self.interval = 3600  # Default: 1 hour in seconds
        self.script_path = Path(__file__).parent / "run.sh"
        self.log_path = Path(__file__).parent / "run.log"
        self.config_path = Path(__file__).parent / "service_config.json"
        self.max_log_lines = 100
        
        # Load configuration if exists
        self.load_config()
        
        # Set up signal handlers
        signal.signal(signal.SIGTERM, self.signal_handler)
        signal.signal(signal.SIGINT, self.signal_handler)
    
    def load_config(self):
        """Load configuration from JSON file"""
        if self.config_path.exists():
            try:
                with open(self.config_path, 'r') as f:
                    config = json.load(f)
                    self.interval = config.get('interval', 3600)
            except Exception as e:
                print(f"Error loading config: {e}")
    
    def save_config(self):
        """Save configuration to JSON file"""
        try:
            config = {'interval': self.interval}
            with open(self.config_path, 'w') as f:
                json.dump(config, f, indent=2)
        except Exception as e:
            print(f"Error saving config: {e}")
    
    def set_interval(self, interval):
        """Set execution interval in seconds"""
        self.interval = interval
        self.save_config()
        print(f"Interval set to {interval} seconds")
    
    def trim_log(self):
        """Keep only the last 100 lines in the log file"""
        if not self.log_path.exists():
            return
        
        try:
            with open(self.log_path, 'r') as f:
                lines = f.readlines()
            
            if len(lines) > self.max_log_lines:
                with open(self.log_path, 'w') as f:
                    f.writelines(lines[-self.max_log_lines:])
        except Exception as e:
            print(f"Error trimming log: {e}")
    
    def execute_script(self):
        """Execute the run.sh script and log results"""
        try:
            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            
            # Execute the script
            result = subprocess.run(
                [str(self.script_path)],
                capture_output=True,
                text=True,
                cwd=self.script_path.parent
            )
            
            # Prepare log entry
            log_entry = f"[{timestamp}] Exit code: {result.returncode}\n"
            if result.stdout:
                log_entry += f"STDOUT: {result.stdout.strip()}\n"
            if result.stderr:
                log_entry += f"STDERR: {result.stderr.strip()}\n"
            log_entry += "-" * 50 + "\n"
            
            # Write to log
            with open(self.log_path, 'a') as f:
                f.write(log_entry)
            
            # Trim log if necessary
            self.trim_log()
            
            print(f"Script executed at {timestamp}, exit code: {result.returncode}")
            
        except Exception as e:
            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            error_msg = f"[{timestamp}] ERROR: {str(e)}\n" + "-" * 50 + "\n"
            
            with open(self.log_path, 'a') as f:
                f.write(error_msg)
            
            print(f"Error executing script: {e}")
    
    def run(self):
        """Main service loop"""
        self.running = True
        print(f"Service started with {self.interval} second interval")
        
        while self.running:
            self.execute_script()
            
            # Sleep in small intervals to allow for graceful shutdown
            elapsed = 0
            while elapsed < self.interval and self.running:
                time.sleep(1)
                elapsed += 1
    
    def stop(self):
        """Stop the service"""
        self.running = False
        print("Service stopping...")
    
    def signal_handler(self, signum, frame):
        """Handle shutdown signals"""
        print(f"Received signal {signum}")
        self.stop()
        sys.exit(0)

def main():
    service = RunScriptService()
    
    if len(sys.argv) > 1:
        command = sys.argv[1]
        
        if command == "set-interval":
            if len(sys.argv) != 3:
                print("Usage: python3 run_script_service.py set-interval <seconds>")
                sys.exit(1)
            try:
                interval = int(sys.argv[2])
                service.set_interval(interval)
            except ValueError:
                print("Invalid interval value. Must be an integer.")
                sys.exit(1)
        elif command == "run":
            service.run()
        else:
            print("Unknown command. Use 'run' or 'set-interval <seconds>'")
            sys.exit(1)
    else:
        service.run()

if __name__ == "__main__":
    main()