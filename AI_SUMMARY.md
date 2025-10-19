# AI Summarization Feature

## การทำงาน

เมื่อค้นหาได้ผลลัพธ์ ระบบจะส่งข้อมูลไปให้ AI สรุปคำตอบอัตโนมัติ:

1. **ลอง Gemini ก่อน** (ฟรี มี limit)
2. **ถ้า Gemini error → ลอง DeepSeek** (fallback)
3. **คืนค่าใน field `summary`**

## ข้อมูลที่ส่งให้ AI

ระบบจะส่งเฉพาะ:
- `content`: เนื้อหาที่ค้นเจอ
- `similarity`: คะแนนความเกี่ยวข้อง

## ตัวอย่าง Response

```json
{
  "query": "ต้องการสมัครงาน",
  "shopid": "shop001",
  "results": [
    {
      "content": "ผู้สมัครงานต้องมีคุณสมบัติดังนี้...",
      "similarity": 0.6194,
      "chunk": 143,
      "filename": "doc01.md"
    }
  ],
  "total": 10,
  "summary": "จากข้อมูลที่ค้นพบ ผู้สมัครงานต้องมีคุณสมบัติดังนี้: 1. มีสัญชาติไทย 2. อายุไม่ต่ำกว่า 18 ปี 3. มีวุฒิการศึกษาตามที่กำหนด..."
}
```

## Environment Variables

ใน `.env`:
```
GEMINI_API_KEY=your-gemini-api-key
DEEPSEEK_API_KEY=your-deepseek-api-key
```

## API ที่ใช้

### Gemini Pro
- URL: `https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent`
- ฟรี แต่มี limit
- ใช้เป็นตัวแรก

### DeepSeek Chat  
- URL: `https://api.deepseek.com/v1/chat/completions`
- Model: `deepseek-chat`
- ใช้เป็น fallback

## การทำงาน

```
Search API
    ↓
ค้นหาข้อมูลจาก Vector DB
    ↓
มีผลลัพธ์? → ใช่
    ↓
ส่งไป Gemini
    ↓
สำเร็จ? → ใช่ → คืนค่า summary
    ↓ ไม่
ส่งไป DeepSeek
    ↓
สำเร็จ? → ใช่ → คืนค่า summary
    ↓ ไม่
คืนค่า summary = ""
```

## Logs

```
🤖 กำลังสรุปผลลัพธ์ด้วย AI...
✅ สรุปผลสำเร็จ
```

หรือ

```
🤖 กำลังสรุปผลลัพธ์ด้วย AI...
⚠️  ไม่สามารถสรุปผลได้
```
