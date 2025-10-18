#!/bin/bash

# Pull Ollama model script
# This script pulls the Qwen3-Embedding-8B model into the Ollama container

echo "Waiting for Ollama service to be ready..."
sleep 10

echo "Pulling Qwen3-Embedding-8B model..."
docker exec ollama ollama pull qwen2.5:0.5b

echo "Verifying installed models..."
docker exec ollama ollama list

echo "Model installation complete!"
