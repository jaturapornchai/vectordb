# Script ตรวจสอบข้อมูลในฐานข้อมูล

Write-Host "=== ตรวจสอบข้อมูลในฐานข้อมูล ===" -ForegroundColor Cyan
Write-Host ""

# อ่านค่าจาก .env
$envFile = Get-Content .env
$dbHost = ($envFile | Where-Object { $_ -match "^DB_HOST=" }) -replace "DB_HOST=", ""
$dbPort = ($envFile | Where-Object { $_ -match "^DB_PORT=" }) -replace "DB_PORT=", ""
$dbUser = ($envFile | Where-Object { $_ -match "^DB_USER=" }) -replace "DB_USER=", ""
$dbPassword = ($envFile | Where-Object { $_ -match "^DB_PASSWORD=" }) -replace "DB_PASSWORD=", ""
$dbName = ($envFile | Where-Object { $_ -match "^DB_NAME=" }) -replace "DB_NAME=", ""

Write-Host "Database: $dbHost`:$dbPort / $dbName" -ForegroundColor Gray
Write-Host ""

# Query 1: จำนวน records ทั้งหมด
Write-Host "1. จำนวน embeddings ทั้งหมด:" -ForegroundColor Yellow
$query1 = "SELECT COUNT(*) as total FROM a;"
$result1 = docker exec -i vectordb-postgres psql -U $dbUser -d $dbName -t -c $query1
Write-Host "   Total: $($result1.Trim()) records" -ForegroundColor Green
Write-Host ""

# Query 2: แยกตาม shopid
Write-Host "2. จำนวน records แยกตาม Shop ID:" -ForegroundColor Yellow
$query2 = "SELECT shopid, COUNT(*) as count FROM a GROUP BY shopid ORDER BY shopid;"
$result2 = docker exec -i vectordb-postgres psql -U $dbUser -d $dbName -t -c $query2
$result2 | ForEach-Object {
    $line = $_.Trim()
    if ($line) {
        Write-Host "   $line" -ForegroundColor Green
    }
}
Write-Host ""

# Query 3: แยกตาม filename
Write-Host "3. จำนวน records แยกตาม Filename:" -ForegroundColor Yellow
$query3 = "SELECT filename, COUNT(*) as count FROM a GROUP BY filename ORDER BY filename;"
$result3 = docker exec -i vectordb-postgres psql -U $dbUser -d $dbName -t -c $query3
$result3 | ForEach-Object {
    $line = $_.Trim()
    if ($line) {
        Write-Host "   $line" -ForegroundColor Green
    }
}
Write-Host ""

# Query 4: แยกตาม shopid และ filename
Write-Host "4. รายละเอียดแยกตาม Shop ID และ Filename:" -ForegroundColor Yellow
$query4 = "SELECT shopid, filename, COUNT(*) as count FROM a GROUP BY shopid, filename ORDER BY shopid, filename;"
$result4 = docker exec -i vectordb-postgres psql -U $dbUser -d $dbName -t -c $query4
$result4 | ForEach-Object {
    $line = $_.Trim()
    if ($line) {
        Write-Host "   $line" -ForegroundColor Green
    }
}
Write-Host ""

Write-Host "=== เสร็จสิ้น ===" -ForegroundColor Cyan
