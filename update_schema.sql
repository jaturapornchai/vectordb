-- Update database schema to support multi-tenant and filename tracking

-- Add new columns to existing table
ALTER TABLE a ADD COLUMN IF NOT EXISTS shopid VARCHAR(50);
ALTER TABLE a ADD COLUMN IF NOT EXISTS filename VARCHAR(255);

-- Update existing records with default values (if any exist)
UPDATE a SET shopid = 'default', filename = source_file WHERE shopid IS NULL;

-- Make shopid and filename NOT NULL after setting defaults
ALTER TABLE a ALTER COLUMN shopid SET NOT NULL;
ALTER TABLE a ALTER COLUMN filename SET NOT NULL;

-- Create composite index for efficient shopid + filename queries
CREATE INDEX IF NOT EXISTS a_shopid_filename_idx ON a(shopid, filename);

-- Create index for shopid searches
CREATE INDEX IF NOT EXISTS a_shopid_idx ON a(shopid);

-- Show current table structure
\d a;