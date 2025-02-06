ALTER TABLE shops DROP COLUMN category;

CREATE TABLE shop_categories(
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    category_member_id INTEGER NOT NULL REFERENCES category_members(id) ON DELETE CASCADE,
    PRIMARY KEY (shop_id, category_member_id)
);