-- Drop the foreign key constraint with ON DELETE CASCADE
ALTER TABLE shops DROP CONSTRAINT shops_created_by_fkey;

-- Restore the original foreign key constraint without ON DELETE CASCADE
ALTER TABLE shops 
ADD CONSTRAINT shops_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(id);
