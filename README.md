# Go Vector Database with PostgreSQL & Ollama

โปรเจค Go application ที่เชื่อมต่อกับ PostgreSQL พร้อม pgvector extension สำหรับการค้นหาแบบ vector similarity และ Ollama สำหรับสร้าง embeddings

## โครงสร้างโปรเจค

```
vectordb/
├── .env                    # การตั้งค่า database connection และ Ollama
├── .github/
│   └── copilot-instructions.md
├── docker-compose.yml      # Docker orchestration (Ollama only)
├── Dockerfile              # Go application container (ไม่ใช้งาน)
├── go.mod                  # Go dependencies
├── main.go                 # Go application หลัก
├── scripts/
│   └── pull-model.sh       # Script สำหรับดาวน์โหลด Ollama model
├── doc/
│   └── doc01.md           # เอกสารตัวอย่างสำหรับสร้าง embeddings
└── README.md
```

## ความต้องการของระบบ

- **Go 1.21+**
- **Docker & Docker Compose** (สำหรับ Ollama)
- **PostgreSQL with pgvector extension** (สามารถใช้ remote server)

## การติดตั้งและเริ่มใช้งาน

### 1. Clone repository

```bash
cd /Users/jaturapornchairatanapanya/git/vectordb
```

### 2. ตรวจสอบไฟล์ `.env`

ไฟล์ `.env` ควรมีการตั้งค่าดังนี้:

```env
DB_HOST=103.13.30.32
DB_PORT=5434
DB_USER=chatbot
DB_PASSWORD=chatbot123
DB_NAME=chatbot
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=qwen2.5:0.5b
```

### 3. เริ่มต้น Ollama Service

```bash
# เริ่ม Ollama container
docker-compose up -d

# ตรวจสอบสถานะ
docker-compose ps
```

### 4. ติดตั้ง Ollama Model (Qwen2.5:0.5b)

```bash
# Pull model (ครั้งแรกเท่านั้น)
docker exec ollama ollama pull qwen2.5:0.5b

# ตรวจสอบ models ที่ติดตั้ง
docker exec ollama ollama list
```

### 5. Download Go dependencies

```bash
go mod download
```

## การใช้งาน

### โหมด 1: สร้าง Vector Database ใหม่ (`-rebuild`)

ใช้เมื่อต้องการสร้างหรืออัพเดท vector database จากเอกสารใน `doc/`:

```bash
go run main.go -rebuild
```

คำสั่งนี้จะ:
1. ✅ สร้างฐานข้อมูล `testvector` (ถ้ายังไม่มี)
2. ✅ สร้างตาราง `a` พร้อม pgvector extension
3. ✅ ล้างข้อมูลเก่าในตาราง
4. ✅ อ่านไฟล์ `.md` ทั้งหมดใน `doc/`
5. ✅ แบ่งเอกสารเป็น chunks (ประมาณ 1000 ตัวอักษรต่อ chunk)
6. ✅ สร้าง embeddings ผ่าน Ollama (qwen2.5:0.5b)
7. ✅ บันทึก embeddings ลงในตาราง `a`

**Output ตัวอย่าง:**
```
=== REBUILD MODE ===
=== Processing documents ===
Found 1 markdown files to process
Processing file: doc/doc01.md
Split doc01.md into 79 chunks
Processing chunk 1/79 from doc01.md
...
Total embeddings in table 'a': 79
```

### โหมด 2: ค้นหาเอกสาร (Search Mode - Default)

ใช้เมื่อต้องการทดสอบการค้นหาด้วย similarity search:

```bash
go run main.go
```

คำสั่งนี้จะ:
1. ✅ เชื่อมต่อกับฐานข้อมูล `testvector`
2. ✅ สุ่มเลือกข้อความจากเอกสารที่บันทึกไว้ (3 ครั้ง)
3. ✅ สร้าง embedding สำหรับคำค้นหา
4. ✅ ค้นหาเอกสารที่คล้ายคลึงที่สุด (Top 3) โดยใช้ cosine similarity
5. ✅ แสดงผลลัพธ์พร้อม similarity score

**Output ตัวอย่าง:**
```
=== SEARCH MODE ===
Database contains 79 embeddings

=== Random Search #1 ===
Query: การลาป่วย พนักงานที่ป่วยและไม่สามารถมาทำงานได้...
Generating embedding for query...

=== Search Results ===
[Result 1] (Similarity: 0.9835)
  File: doc01.md, Chunk: 11
  Content: - วันทำงานปกติ อัตรา 1.5 เท่าของค่าจ้างต่อชั่วโมง...
```

## โครงสร้าง Database

### ฐานข้อมูล: `testvector`

### ตาราง: `a`

```sql
CREATE TABLE a (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,              -- เนื้อหาของ chunk
    chunk_index INTEGER NOT NULL,       -- ลำดับของ chunk ในเอกสาร
    source_file TEXT NOT NULL,          -- ชื่อไฟล์ต้นทาง
    embedding vector(896),              -- Vector embedding (896 มิติ)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index สำหรับ similarity search
CREATE INDEX a_embedding_idx 
ON a USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

## การทำงานของ Similarity Search

Application ใช้ **Cosine Similarity** สำหรับค้นหาเอกสารที่คล้ายคลึงกัน:

```sql
SELECT id, source_file, chunk_index, content,
       1 - (embedding <=> '[query_vector]'::vector) as similarity
FROM a
WHERE embedding IS NOT NULL
ORDER BY embedding <=> '[query_vector]'::vector
LIMIT 3
```

- `<=>` = cosine distance operator ของ pgvector
- Similarity = 1.0 (เหมือนกันทุกประการ)
- Similarity = 0.0 (แตกต่างกันโดยสิ้นเชิง)

## คำสั่งที่เป็นประโยชน์

### Docker (Ollama)

```bash
# เริ่ม Ollama
docker-compose up -d

# หยุด Ollama
docker-compose down

# ดู logs
docker-compose logs -f ollama

# ทดสอบ Ollama API
curl http://localhost:11434/api/tags
```

### Go Application

```bash
# สร้าง vector database ใหม่
go run main.go -rebuild

# ค้นหาเอกสาร (โหมดปกติ)
go run main.go

# Build executable
go build -o vectordb main.go

# รัน executable
./vectordb              # Search mode
./vectordb -rebuild     # Rebuild mode
```

### Database

```bash
# เชื่อมต่อ PostgreSQL
psql -h 103.13.30.32 -p 5434 -U chatbot -d testvector

# ตรวจสอบข้อมูล
SELECT COUNT(*) FROM a;
SELECT source_file, COUNT(*) as chunks FROM a GROUP BY source_file;

# ดูตัวอย่างข้อมูล
SELECT id, source_file, chunk_index, LEFT(content, 100) 
FROM a ORDER BY id LIMIT 5;
```

## Architecture

```
┌─────────────────┐
│   Go App        │
│  (Local)        │
└────┬────────┬───┘
     │        │
     │        └──────────────┐
     │                       │
     ▼                       ▼
┌─────────────┐      ┌──────────────┐
│  Ollama     │      │  PostgreSQL  │
│  (Docker)   │      │  + pgvector  │
│  :11434     │      │  (Remote)    │
└─────────────┘      └──────────────┘
```

- **Go Application**: รันบนเครื่อง local
- **Ollama**: รันบน Docker container (localhost:11434)
- **PostgreSQL**: เชื่อมต่อกับ remote server พร้อม pgvector extension

## Technologies

- **Go 1.21**: Programming language
- **PostgreSQL + pgvector**: Vector database
- **Ollama (qwen2.5:0.5b)**: Embedding generation
- **Docker & Docker Compose**: Containerization

## ตัวอย่างการใช้งาน

### 1. สร้าง Vector Database ครั้งแรก

```bash
# เริ่ม Ollama
docker-compose up -d

# รอให้ Ollama พร้อม
sleep 5

# Pull model (ถ้ายังไม่มี)
docker exec ollama ollama pull qwen2.5:0.5b

# สร้าง vector database
go run main.go -rebuild
```

### 2. ค้นหาเอกสาร

```bash
# รันในโหมดค้นหา
go run main.go

# หรือ
./vectordb
```

### 3. เพิ่มเอกสารใหม่

```bash
# 1. เพิ่มไฟล์ .md ใหม่ใน doc/
# 2. Rebuild database
go run main.go -rebuild
```

## Troubleshooting

### ไม่สามารถเชื่อมต่อ Database

ตรวจสอบว่า PostgreSQL server ที่ `103.13.30.32:5434` สามารถเข้าถึงได้และมี pgvector extension:

```sql
-- ตรวจสอบ extension
SELECT * FROM pg_extension WHERE extname = 'vector';

-- สร้าง extension (ถ้ายังไม่มี)
CREATE EXTENSION vector;
```

### Ollama ไม่ตอบสนอง

```bash
# ตรวจสอบ container
docker ps

# ดู logs
docker-compose logs ollama

# Restart
docker-compose restart ollama

# ทดสอบ API
curl http://localhost:11434/api/tags
```

### Database ว่างเปล่า

```bash
# ตรวจสอบว่ามีข้อมูลหรือไม่
psql -h 103.13.30.32 -p 5434 -U chatbot -d testvector -c "SELECT COUNT(*) FROM a;"

# ถ้าว่าง ให้ rebuild
go run main.go -rebuild
```

## Performance Tips

1. **Chunk Size**: ปรับขนาด chunk ใน `chunkText()` function (ปัจจุบัน: 1000 ตัวอักษร)
2. **Index**: ปรับ `lists` parameter ใน ivfflat index ตามขนาดข้อมูล
3. **Batch Processing**: เพิ่ม batch insert สำหรับข้อมูลจำนวนมาก
4. **Caching**: cache embeddings ที่ใช้บ่อยๆ

## License

MIT

## ผู้พัฒนา

Jaturaporn Chairatanapanya
