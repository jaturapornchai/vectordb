-- Drop old table and create new with 768 dimensions for nomic-embed-text
DROP TABLE IF EXISTS a;

CREATE TABLE a (
    id SERIAL PRIMARY KEY,
    content TEXT,
    source_file TEXT,
    filename TEXT,
    shopid TEXT,
    chunk_index INTEGER,
    embedding vector(768),  -- Changed from 1024 to 768
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster vector search
CREATE INDEX ON a USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create index for shopid filtering
CREATE INDEX idx_shopid ON a(shopid);
CREATE INDEX idx_filename ON a(filename);
CREATE INDEX idx_shopid_filename ON a(shopid, filename);
