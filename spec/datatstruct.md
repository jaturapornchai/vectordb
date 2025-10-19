-- ========================================
-- Vector Database Schema
-- ========================================

-- เปิดใช้งาน pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- ========================================
-- Table: documents
-- เก็บ embeddings และข้อมูลเอกสาร
-- ========================================
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    shopid VARCHAR(50) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1024),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- สร้าง index สำหรับการค้นหา vector
CREATE INDEX IF NOT EXISTS documents_embedding_idx ON documents 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- สร้าง index สำหรับการค้นหาด้วย shopid และ filename
CREATE INDEX IF NOT EXISTS documents_shopid_idx ON documents(shopid);
CREATE INDEX IF NOT EXISTS documents_filename_idx ON documents(filename);
CREATE INDEX IF NOT EXISTS documents_shopid_filename_idx ON documents(shopid, filename);

-- ========================================
-- Table: shopidfilename
-- เก็บประวัติการสร้าง embeddings
-- ========================================
CREATE TABLE IF NOT EXISTS shopidfilename (
    id SERIAL PRIMARY KEY,
    shopid VARCHAR(50) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    emailusercreate VARCHAR(255),
    createdate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shopid, filename)
);

-- สร้าง index สำหรับการค้นหา
CREATE INDEX IF NOT EXISTS shopidfilename_shopid_idx ON shopidfilename(shopid);
CREATE INDEX IF NOT EXISTS shopidfilename_filename_idx ON shopidfilename(filename);

-- สร้าง trigger สำหรับอัปเดต updatedate อัตโนมัติ
CREATE OR REPLACE FUNCTION update_updatedate_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedate = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_shopidfilename_updatedate 
    BEFORE UPDATE ON shopidfilename 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updatedate_column();

-- ========================================
-- คำสั่งสำหรับตรวจสอบข้อมูล
-- ========================================

-- ดูจำนวน embeddings แต่ละ shop
-- SELECT shopid, filename, COUNT(*) as total_embeddings 
-- FROM documents 
-- GROUP BY shopid, filename 
-- ORDER BY shopid, filename;

-- ดูประวัติการสร้าง embeddings
-- SELECT * FROM shopidfilename ORDER BY createdate DESC;

-- ลบข้อมูลทั้งหมด (ระวัง!)
-- TRUNCATE TABLE documents CASCADE;
-- TRUNCATE TABLE shopidfilename CASCADE;
