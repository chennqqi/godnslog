#!/bin/bash
# Build script for GODNSLOG Burp Extension

echo "Building GODNSLOG Burp Extension..."

# Check if Maven is installed
if ! command -v mvn &> /dev/null; then
    echo "Maven is not installed. Please install Maven first."
    exit 1
fi

# Build the project
mvn clean package

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "JAR file created in target/godnslog-burp-2.0.0.jar"
else
    echo "Build failed!"
    exit 1
fi
