#!/bin/bash
# Quick Start Script - Mac/Linux Desktop
# ใช้สำหรับ deploy บน Mac/Linux Desktop ครั้งแรก

echo "🚀 Vector Search API - Desktop Deployment"
echo "========================================"
echo ""

# 1. ตรวจสอบ Docker
echo "1️⃣  ตรวจสอบ Docker..."
if ! docker ps > /dev/null 2>&1; then
    echo "❌ Docker ไม่ทำงาน! กรุณาเปิด Docker ก่อน"
    exit 1
fi
echo "   ✅ Docker ทำงานอยู่"
echo ""

# 2. ตรวจสอบไฟล์ที่จำเป็น
echo "2️⃣  ตรวจสอบไฟล์ที่จำเป็น..."
REQUIRED_FILES=("docker-compose.yml" "Dockerfile" ".env" "main.go")
MISSING_FILES=()

for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        MISSING_FILES+=("$file")
    fi
done

if [ ${#MISSING_FILES[@]} -ne 0 ]; then
    echo "   ❌ ไฟล์ต่อไปนี้ยังไม่มี: ${MISSING_FILES[*]}"
    if [[ " ${MISSING_FILES[@]} " =~ " .env " ]]; then
        echo "   💡 กรุณาสร้างไฟล์ .env และใส่ API keys"
    fi
    exit 1
fi
echo "   ✅ ไฟล์ครบถ้วน"
echo ""

# 3. ตรวจสอบโฟลเดอร์เอกสาร
echo "3️⃣  ตรวจสอบเอกสาร..."
if [ ! -d "doc" ]; then
    echo "   ⚠️  ไม่พบโฟลเดอร์ doc/ กำลังสร้าง..."
    mkdir -p doc
fi
DOC_COUNT=$(find doc -name "*.md" 2>/dev/null | wc -l)
echo "   ✅ พบเอกสาร $DOC_COUNT ไฟล์"
echo ""

# 4. Stop services เดิม (ถ้ามี)
echo "4️⃣  หยุด services เดิม..."
docker-compose down > /dev/null 2>&1
echo "   ✅ เสร็จสิ้น"
echo ""

# 5. Build Docker images
echo "5️⃣  Build Docker images..."
docker-compose build
if [ $? -ne 0 ]; then
    echo "   ❌ Build ไม่สำเร็จ!"
    exit 1
fi
echo "   ✅ Build สำเร็จ"
echo ""

# 6. Start services
echo "6️⃣  Start services..."
docker-compose up -d
if [ $? -ne 0 ]; then
    echo "   ❌ Start ไม่สำเร็จ!"
    exit 1
fi
echo "   ✅ Services เริ่มทำงานแล้ว"
echo ""

# 7. รอให้ Ollama พร้อม
echo "7️⃣  รอ Ollama เริ่มต้น (10 วินาที)..."
sleep 10
echo "   ✅ เสร็จสิ้น"
echo ""

# 8. ตรวจสอบว่ามี model llama3.2 แล้วหรือยัง
echo "8️⃣  ตรวจสอบ Ollama models..."
MODELS=$(docker exec ollama ollama list 2>/dev/null)
if ! echo "$MODELS" | grep -q "llama3.2"; then
    echo "   ⚠️  ยังไม่มี llama3.2 model กำลัง download..."
    echo "   📥 (ขนาด ~2GB, ใช้เวลา 5-10 นาที)"
    echo ""
    
    docker exec ollama ollama pull llama3.2:latest
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "   ✅ Download สำเร็จ"
    else
        echo ""
        echo "   ❌ Download ไม่สำเร็จ"
        exit 1
    fi
else
    echo "   ✅ มี llama3.2 model แล้ว"
fi
echo ""

# 9. Test API
echo "9️⃣  ทดสอบ API..."
sleep 3

HEALTH=$(curl -s http://localhost:8080/health)
if [ $? -eq 0 ]; then
    echo "   ✅ API ทำงานปกติ"
    echo "   📊 Response: $HEALTH"
else
    echo "   ❌ API ยังไม่พร้อม รอสักครู่แล้วลองใหม่"
    echo "   💡 ใช้คำสั่ง: docker-compose logs -f vectordb-api"
fi
echo ""

# 10. แสดงข้อมูลสรุป
echo "========================================"
echo "✅ Deployment เสร็จสมบูรณ์!"
echo "========================================"
echo ""

echo "📍 API Endpoint:"
echo "   http://localhost:8080/search"
echo ""

echo "🔧 คำสั่งที่มีประโยชน์:"
echo "   ดู logs:        docker-compose logs -f vectordb-api"
echo "   Stop services:  docker-compose stop"
echo "   Start services: docker-compose start"
echo "   Restart:        docker-compose restart"
echo "   ลบทั้งหมด:      docker-compose down -v"
echo ""

echo "🧪 ทดสอบ API:"
echo '   curl -X POST http://localhost:8080/search \'
echo '     -H "Content-Type: application/json" \'
echo '     -d '"'"'{"query":"กระเบื้อง","useSummary":false}'"'"
echo ""

echo "📚 คู่มือเพิ่มเติม: DEPLOY_DESKTOP.md"
echo ""

echo "🎉 พร้อมใช้งาน!"
