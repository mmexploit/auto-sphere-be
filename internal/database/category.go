package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type CategoryModel struct {
	db *sql.DB
}

type Category struct {
	Id    int
	Value string
}

func (cm CategoryModel) Create(cat *Category) error {

	query := `INSERT INTO categories (value)
			VALUES ($1)
			RETURNING id, value`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	args := []interface{}{
		cat.Value,
	}
	return cm.db.QueryRowContext(ctx, query, args...).Scan(&cat.Id, &cat.Value)
}

func (cm CategoryModel) Put(cat *Category) error {
	query := `UPDATE categories SET value=$1
			WHERE id=$2 
			RETURNING id, value`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	args := []interface{}{
		cat.Value,
		cat.Id,
	}

	return cm.db.QueryRowContext(ctx, query, args...).Scan(&cat.Id, &cat.Value)
}
func (cm CategoryModel) Get(id int64) (*Category, error) {

	query := `Select id, value FROM categories
			  WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	var cat Category
	err := cm.db.QueryRowContext(ctx, query, id).Scan(&cat.Id, &cat.Value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &cat, nil
}

func (cm CategoryModel) GetAll() ([]Category, int, error) {

	query := `SELECT count(*) OVER(), id, value FROM categories;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := cm.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrRecordNotFound
		}
		return nil, 0, err
	}

	var categories []Category
	totalAmount := 0
	for rows.Next() {
		var category Category
		err := rows.Scan(
			&totalAmount,
			&category.Id,
			&category.Value,
		)
		if err != nil {
			return nil, 0, err
		}
		categories = append(categories, category)
	}
	return categories, totalAmount, nil
}

func (cm CategoryModel) Delete(id int64) error {
	query := `DELETE FROM categories WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := cm.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
