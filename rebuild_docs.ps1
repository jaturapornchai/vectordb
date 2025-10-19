# Script สำหรับ Rebuild เอกสารทั้งหมดใน doc/ folder
# ใช้ chunk size 600 characters

$apiUrl = "http://localhost:8080/build"
$shopId = "shop001"
$docFolder = ".\doc"

Write-Host "🚀 เริ่มต้น Rebuild เอกสารทั้งหมด..." -ForegroundColor Green
Write-Host "📁 โฟลเดอร์: $docFolder" -ForegroundColor Cyan
Write-Host "🏪 Shop ID: $shopId" -ForegroundColor Cyan
Write-Host "📏 Chunk Size: 600 characters" -ForegroundColor Cyan
Write-Host ""

# ค้นหาไฟล์ .md ทั้งหมด
$mdFiles = Get-ChildItem -Path $docFolder -Filter *.md

if ($mdFiles.Count -eq 0) {
    Write-Host "❌ ไม่พบไฟล์ .md ใน folder $docFolder" -ForegroundColor Red
    exit
}

Write-Host "📄 พบไฟล์ทั้งหมด: $($mdFiles.Count) ไฟล์" -ForegroundColor Yellow
Write-Host ""

$successCount = 0
$errorCount = 0

foreach ($file in $mdFiles) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host "📝 กำลังประมวลผล: $($file.Name)" -ForegroundColor White
    
    try {
        # อ่านเนื้อหาไฟล์
        $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
        
        Write-Host "   📊 ขนาดไฟล์: $($content.Length) ตัวอักษร" -ForegroundColor Gray
        
        # สร้าง JSON body
        $body = @{
            filename = $file.Name
            shopid = $shopId
            content = $content
        } | ConvertTo-Json -Depth 10
        
        # ส่ง request
        Write-Host "   ⏳ กำลังส่งข้อมูลไป API..." -ForegroundColor Yellow
        
        $response = Invoke-RestMethod -Uri $apiUrl `
            -Method POST `
            -ContentType "application/json; charset=utf-8" `
            -Body $body `
            -TimeoutSec 300
        
        if ($response.status -eq "success") {
            Write-Host "   ✅ สำเร็จ! สร้าง $($response.chunks) chunks" -ForegroundColor Green
            $successCount++
        } else {
            Write-Host "   ❌ ล้มเหลว: $($response.error)" -ForegroundColor Red
            $errorCount++
        }
        
    } catch {
        Write-Host "   [ERROR] Exception: $($_.Exception.Message)" -ForegroundColor Red
        $errorCount++
    }
    
    Write-Host ""
    Start-Sleep -Milliseconds 500
}

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host ""
Write-Host "🎉 สรุปผลการ Rebuild:" -ForegroundColor Cyan
Write-Host "   ✅ สำเร็จ: $successCount ไฟล์" -ForegroundColor Green
Write-Host "   ❌ ล้มเหลว: $errorCount ไฟล์" -ForegroundColor Red
Write-Host "   📊 รวมทั้งหมด: $($mdFiles.Count) ไฟล์" -ForegroundColor White
Write-Host ""

if ($errorCount -eq 0) {
    Write-Host "✨ Rebuild เสร็จสมบูรณ์! ระบบพร้อมใช้งาน" -ForegroundColor Green
} else {
    Write-Host "⚠️  มีไฟล์บางไฟล์ที่ไม่สำเร็จ กรุณาตรวจสอบ" -ForegroundColor Yellow
}
