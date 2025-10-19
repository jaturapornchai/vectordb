# Pull Llama3.2 model for query expansion (PowerShell)

Write-Host "🚀 กำลังดึง Llama3.2 model..." -ForegroundColor Cyan
docker exec ollama ollama pull llama3.2:latest

Write-Host "`n✅ เสร็จแล้ว! ตรวจสอบ models:" -ForegroundColor Green
docker exec ollama ollama list
