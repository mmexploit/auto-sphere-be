ALTER TABLE category_members 
ADD COLUMN category_id INT NOT NULL,
ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;