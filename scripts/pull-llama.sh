#!/bin/bash
# Pull Llama3.2 model for query expansion

echo "🚀 กำลังดึง Llama3.2 model..."
docker exec -it ollama ollama pull llama3.2:latest

echo "✅ เสร็จแล้ว! ตรวจสอบ models:"
docker exec -it ollama ollama list
