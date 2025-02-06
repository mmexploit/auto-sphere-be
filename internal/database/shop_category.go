package database

import (
	"context"
	"database/sql"
	"time"
)

type ShopCategoryModel struct {
	db *sql.DB
}

type ShopCategory struct {
	Shop_id            int
	Category_member_id int
}

func (scm ShopCategoryModel) Create(sc ShopCategory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `INSERT INTO shop_categories (shop_id, category_member_id)
			  VALUES($1,$2)
			  RETURNING shop_id, category_member_id`
	args := []interface{}{sc.Shop_id, sc.Category_member_id}
	return scm.db.QueryRowContext(ctx, query, args...).Scan(&sc.Shop_id, &sc.Category_member_id)
}

func (scm ShopCategoryModel) Delete(shop_id int, category_member_id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM shop_categories
			  WHERE shop_id=$1 AND category_member_id=$2`
	_, err := scm.db.ExecContext(ctx, query, shop_id, category_member_id)
	return err
}
