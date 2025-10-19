-- สร้าง table เก็บประวัติการสร้าง embeddings
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
CREATE INDEX IF NOT EXISTS shopidfilename_createdate_idx ON shopidfilename(createdate);

-- Function สำหรับอัปเดต updatedate อัตโนมัติ
CREATE OR REPLACE FUNCTION update_shopidfilename_updatedate()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedate = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- สร้าง trigger สำหรับอัปเดต updatedate
DROP TRIGGER IF EXISTS shopidfilename_updatedate_trigger ON shopidfilename;
CREATE TRIGGER shopidfilename_updatedate_trigger
    BEFORE UPDATE ON shopidfilename
    FOR EACH ROW
    EXECUTE FUNCTION update_shopidfilename_updatedate();

-- แสดงโครงสร้างตาราง
\d shopidfilename;
