# Quick Start Script - Windows Desktop
# ใช้สำหรับ deploy บน Windows Desktop ครั้งแรก

Write-Host "🚀 Vector Search API - Desktop Deployment" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. ตรวจสอบ Docker
Write-Host "1️⃣  ตรวจสอบ Docker Desktop..." -ForegroundColor Yellow
$dockerRunning = docker ps 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker Desktop ไม่ทำงาน! กรุณาเปิด Docker Desktop ก่อน" -ForegroundColor Red
    exit 1
}
Write-Host "   ✅ Docker Desktop ทำงานอยู่`n" -ForegroundColor Green

# 2. ตรวจสอบไฟล์ที่จำเป็น
Write-Host "2️⃣  ตรวจสอบไฟล์ที่จำเป็น..." -ForegroundColor Yellow
$requiredFiles = @("docker-compose.yml", "Dockerfile", ".env", "main.go")
$missingFiles = @()

foreach ($file in $requiredFiles) {
    if (-Not (Test-Path $file)) {
        $missingFiles += $file
    }
}

if ($missingFiles.Count -gt 0) {
    Write-Host "   ❌ ไฟล์ต่อไปนี้ยังไม่มี: $($missingFiles -join ', ')" -ForegroundColor Red
    if ($missingFiles -contains ".env") {
        Write-Host "   💡 กรุณาสร้างไฟล์ .env และใส่ API keys" -ForegroundColor Yellow
    }
    exit 1
}
Write-Host "   ✅ ไฟล์ครบถ้วน`n" -ForegroundColor Green

# 3. ตรวจสอบโฟลเดอร์เอกสาร
Write-Host "3️⃣  ตรวจสอบเอกสาร..." -ForegroundColor Yellow
if (-Not (Test-Path "doc")) {
    Write-Host "   ⚠️  ไม่พบโฟลเดอร์ doc/ กำลังสร้าง..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path "doc" -Force | Out-Null
}
$docFiles = Get-ChildItem -Path "doc" -Filter "*.md" -ErrorAction SilentlyContinue
Write-Host "   ✅ พบเอกสาร $($docFiles.Count) ไฟล์`n" -ForegroundColor Green

# 4. Stop services เดิม (ถ้ามี)
Write-Host "4️⃣  หยุด services เดิม..." -ForegroundColor Yellow
docker-compose down 2>$null | Out-Null
Write-Host "   ✅ เสร็จสิ้น`n" -ForegroundColor Green

# 5. Build Docker images
Write-Host "5️⃣  Build Docker images..." -ForegroundColor Yellow
docker-compose build
if ($LASTEXITCODE -ne 0) {
    Write-Host "   ❌ Build ไม่สำเร็จ!" -ForegroundColor Red
    exit 1
}
Write-Host "   ✅ Build สำเร็จ`n" -ForegroundColor Green

# 6. Start services
Write-Host "6️⃣  Start services..." -ForegroundColor Yellow
docker-compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "   ❌ Start ไม่สำเร็จ!" -ForegroundColor Red
    exit 1
}
Write-Host "   ✅ Services เริ่มทำงานแล้ว`n" -ForegroundColor Green

# 7. รอให้ Ollama พร้อม
Write-Host "7️⃣  รอ Ollama เริ่มต้น (10 วินาที)..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
Write-Host "   ✅ เสร็จสิ้น`n" -ForegroundColor Green

# 8. ตรวจสอบว่ามี model llama3.2 แล้วหรือยัง
Write-Host "8️⃣  ตรวจสอบ Ollama models..." -ForegroundColor Yellow
$models = docker exec ollama ollama list 2>$null
if ($models -notmatch "llama3.2") {
    Write-Host "   ⚠️  ยังไม่มี llama3.2 model กำลัง download..." -ForegroundColor Yellow
    Write-Host "   📥 (ขนาด ~2GB, ใช้เวลา 5-10 นาที)`n" -ForegroundColor Yellow
    
    docker exec ollama ollama pull llama3.2:latest
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n   ✅ Download สำเร็จ`n" -ForegroundColor Green
    } else {
        Write-Host "`n   ❌ Download ไม่สำเร็จ" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "   ✅ มี llama3.2 model แล้ว`n" -ForegroundColor Green
}

# 9. Test API
Write-Host "9️⃣  ทดสอบ API..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

try {
    $healthCheck = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -ErrorAction Stop
    Write-Host "   ✅ API ทำงานปกติ" -ForegroundColor Green
    Write-Host "   📊 Status: $($healthCheck.status)" -ForegroundColor Cyan
    Write-Host "   📝 Message: $($healthCheck.message)`n" -ForegroundColor Cyan
} catch {
    Write-Host "   ❌ API ยังไม่พร้อม รอสักครู่แล้วลองใหม่" -ForegroundColor Red
    Write-Host "   💡 ใช้คำสั่ง: docker-compose logs -f vectordb-api`n" -ForegroundColor Yellow
}

# 10. แสดงข้อมูลสรุป
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ Deployment เสร็จสมบูรณ์!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "📍 API Endpoint:" -ForegroundColor Yellow
Write-Host "   http://localhost:8080/search`n" -ForegroundColor White

Write-Host "🔧 คำสั่งที่มีประโยชน์:" -ForegroundColor Yellow
Write-Host "   ดู logs:        docker-compose logs -f vectordb-api" -ForegroundColor White
Write-Host "   Stop services:  docker-compose stop" -ForegroundColor White
Write-Host "   Start services: docker-compose start" -ForegroundColor White
Write-Host "   Restart:        docker-compose restart" -ForegroundColor White
Write-Host "   ลบทั้งหมด:      docker-compose down -v`n" -ForegroundColor White

Write-Host "🧪 ทดสอบ API:" -ForegroundColor Yellow
Write-Host '   $body = @{ query = "กระเบื้อง"; useSummary = $false } | ConvertTo-Json' -ForegroundColor White
Write-Host '   Invoke-RestMethod -Uri "http://localhost:8080/search" -Method Post -Body $body -ContentType "application/json; charset=utf-8"' -ForegroundColor White
Write-Host ""

Write-Host "📚 คู่มือเพิ่มเติม: DEPLOY_DESKTOP.md`n" -ForegroundColor Yellow

Write-Host "🎉 พร้อมใช้งาน!" -ForegroundColor Green
