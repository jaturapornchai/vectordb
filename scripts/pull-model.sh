#!/bin/bash

# Pull Ollama model script
# This script pulls the BGE-M3 model into the Ollama container

echo "Waiting for Ollama service to be ready..."
sleep 10

echo "Pulling BGE-M3 model..."
docker exec ollama ollama pull bge-m3:latest

echo "Verifying installed models..."
docker exec ollama ollama list

echo "Model installation complete!"
