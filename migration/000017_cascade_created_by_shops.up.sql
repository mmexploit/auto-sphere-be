-- Drop the existing foreign key constraint
ALTER TABLE shops DROP CONSTRAINT shops_created_by_fkey;

-- Add the new foreign key constraint with ON DELETE CASCADE
ALTER TABLE shops 
ADD CONSTRAINT shops_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;
