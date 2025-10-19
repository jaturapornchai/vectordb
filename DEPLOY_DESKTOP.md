# Vector Search API - Desktop Deployment Guide

## 📦 สิ่งที่ต้องเตรียม

1. **Docker Desktop** (Windows/Mac/Linux)
   - Download: https://www.docker.com/products/docker-desktop
   - ต้องเปิด Docker Desktop ทุกครั้งก่อนใช้งาน

2. **ไฟล์ที่ต้องมี**
   - `docker-compose.yml`
   - `Dockerfile`
   - `.env` (API keys)
   - `doc/` folder (เอกสารที่จะค้นหา)
   - Source code (*.go files)

## 🚀 วิธี Deploy บน Desktop

### 1. Clone หรือ Copy โปรเจค

```powershell
# ถ้ามี git
git clone <repository-url>
cd vectordb

# หรือ copy folder ทั้งหมดไปที่ Desktop
```

### 2. เตรียม .env File

สร้างไฟล์ `.env` ในโฟลเดอร์ vectordb:

```env
# AI API Configuration
GEMINI_API_KEY=your_gemini_api_key_here
DEEPSEEK_API_KEY=your_deepseek_api_key_here

# Ollama Configuration
OLLAMA_HOST=http://ollama:11434
OLLAMA_MODEL=llama3.2

# Database (ไม่จำเป็นสำหรับ text search)
DB_HOST=localhost
DB_PORT=5432
```

### 3. Pull Ollama Model ครั้งแรก

**Windows (PowerShell):**
```powershell
# Start services
docker-compose up -d

# รอ 10 วินาที ให้ Ollama เริ่มต้น
Start-Sleep -Seconds 10

# Pull model
docker exec ollama ollama pull llama3.2:latest

# ตรวจสอบ
docker exec ollama ollama list
```

**Mac/Linux (Terminal):**
```bash
# Start services
docker-compose up -d

# รอ 10 วินาที
sleep 10

# Pull model
docker exec ollama ollama pull llama3.2:latest

# ตรวจสอบ
docker exec ollama ollama list
```

### 4. Start Services

```powershell
# Windows
docker-compose up -d

# Mac/Linux
docker-compose up -d
```

### 5. ตรวจสอบสถานะ

```powershell
# ดู logs
docker-compose logs -f vectordb-api

# ตรวจสอบว่า service ทำงาน
docker ps

# ทดสอบ health check
curl http://localhost:8080/health
```

## 📝 วิธีใช้งาน

### ค้นหาเอกสาร (ไม่มีสรุป)

**Windows PowerShell:**
```powershell
$body = @{
    query = "กระเบื้อง"
    useSummary = $false
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/search" `
    -Method Post `
    -Body $body `
    -ContentType "application/json; charset=utf-8"
```

**Mac/Linux (curl):**
```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "กระเบื้อง",
    "useSummary": false
  }'
```

### ค้นหาเอกสาร + สรุปด้วย AI

**Windows PowerShell:**
```powershell
$body = @{
    query = "กระเบื้อง"
    useSummary = $true
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/search" `
    -Method Post `
    -Body $body `
    -ContentType "application/json; charset=utf-8"

Write-Host "`n=== สรุป ===" -ForegroundColor Cyan
Write-Host $response.summary -ForegroundColor Green
Write-Host "`n=== พบ $($response.total) ผลลัพธ์ ===" -ForegroundColor Yellow
```

**Mac/Linux (curl + jq):**
```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "กระเบื้อง",
    "useSummary": true
  }' | jq '.summary'
```

## 🛠️ จัดการ Services

### Stop Services
```powershell
docker-compose stop
```

### Restart Services
```powershell
docker-compose restart
```

### ลบทั้งหมด (รวม volumes)
```powershell
docker-compose down -v
```

### Rebuild ใหม่หลังแก้ code
```powershell
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 📂 เพิ่มเอกสารใหม่

1. วาง markdown files ในโฟลเดอร์ `doc/`
2. Restart service:
   ```powershell
   docker-compose restart vectordb-api
   ```
3. เอกสารใหม่จะถูกค้นหาทันที

## 🔍 ตัวอย่างการใช้งาน

### 1. ค้นหาภาษาไทย
```powershell
# ค้นหา "กระเบื้อง"
$body = @{ query = "กระเบื้อง"; useSummary = $false } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/search" -Method Post -Body $body -ContentType "application/json; charset=utf-8"
```

### 2. ค้นหาภาษาอังกฤษ (จะแปลเป็นไทยอัตโนมัติ)
```powershell
# ค้นหา "tile" → Ollama จะแปลเป็น "กระเบื้อง"
$body = @{ query = "tile"; useSummary = $false } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/search" -Method Post -Body $body -ContentType "application/json; charset=utf-8"
```

### 3. ค้นหา + สรุป
```powershell
$body = @{ query = "โปรโมชั่น"; useSummary = $true } | ConvertTo-Json
$result = Invoke-RestMethod -Uri "http://localhost:8080/search" -Method Post -Body $body -ContentType "application/json; charset=utf-8"
$result.summary
```

## ❓ Troubleshooting

### Port 8080 ถูกใช้แล้ว
แก้ไขใน `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # เปลี่ยนเป็น port อื่น
```

### Ollama ใช้ RAM เยอะ
แก้ไขใน `docker-compose.yml`:
```yaml
ollama:
  deploy:
    resources:
      limits:
        memory: 2G  # จำกัด RAM
```

### ไม่พบเอกสาร
```powershell
# ตรวจสอบว่ามีไฟล์ใน container
docker exec vectordb-api ls -la /app/doc/

# ถ้าไม่มี rebuild
docker-compose down
docker-compose build
docker-compose up -d
```

### Ollama ช้า
```powershell
# ใช้ model เล็กกว่า
docker exec ollama ollama pull llama3.2:1b

# แก้ไขใน queryexpansion.go
Model: "llama3.2:1b"
```

## 📊 Performance

- **RAM**: ≥ 8GB แนะนำ
- **Disk**: ≥ 5GB สำหรับ Ollama models
- **CPU**: ≥ 4 cores แนะนำ

**Ollama Models:**
- `llama3.2:1b` - เล็กที่สุด, เร็วที่สุด (1.3GB)
- `llama3.2:3b` - กลาง (2GB)
- `llama3.2:latest` - ใหญ่, แม่นที่สุด (2GB)

## 🔐 Security

1. **อย่า commit `.env` file** (ใส่ใน `.gitignore`)
2. **API Keys**: เก็บใน `.env` เท่านั้น
3. **Production**: ใช้ reverse proxy (nginx) + SSL

## 📱 Integration

### Python
```python
import requests

response = requests.post(
    "http://localhost:8080/search",
    json={"query": "กระเบื้อง", "useSummary": True}
)
print(response.json()["summary"])
```

### Node.js
```javascript
const axios = require('axios');

axios.post('http://localhost:8080/search', {
  query: 'กระเบื้อง',
  useSummary: true
}).then(res => {
  console.log(res.data.summary);
});
```

### cURL
```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query":"กระเบื้อง","useSummary":true}'
```

## 🎯 Features

✅ **ค้นหาด้วย AI**: ขยายคำค้นหาอัตโนมัติ (คำพ้องเสียง, แปลภาษา)
✅ **Multi-language**: รองรับไทย-อังกฤษ
✅ **Smart Summary**: สรุปผลด้วย Gemini/DeepSeek API
✅ **Fast**: Text search แบบ realtime
✅ **No Database**: ไม่ต้องติดตั้ง PostgreSQL
✅ **Portable**: ใช้ได้บน Windows/Mac/Linux

## 📞 Support

- GitHub Issues: [repository-url]/issues
- Email: your-email@example.com
