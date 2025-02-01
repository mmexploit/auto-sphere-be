package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CategoryMemberModel struct {
	db *sql.DB
}

type CategoryMember struct {
	Id         int    `json:"id"`
	Value      string `json:"value"`
	CategoryId int    `json:"category_id"`
}

func (cmm CategoryMemberModel) Create(cm *CategoryMember) error {
	fmt.Print("The CM !! ", cm.CategoryId, "  !!")
	query := `INSERT INTO category_members (value, category_id)
			VALUES ($1,$2)
			RETURNING id, value, category_id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{cm.Value, cm.CategoryId}
	return cmm.db.QueryRowContext(ctx, query, args...).Scan(&cm.Id, &cm.Value, &cm.CategoryId)
}

func (cmm CategoryMemberModel) Patch(cm *CategoryMember) error {
	query := `UPDATE category_members 
			SET value=$1, category_id=$2 
			WHERE id=$3 
			RETURNING id, value, category_id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{cm.Value, cm.CategoryId, cm.Id}
	return cmm.db.QueryRowContext(ctx, query, args...).Scan(&cm.Id, &cm.Value, &cm.CategoryId)
}

func (cmm CategoryMemberModel) Get(id int64) (*CategoryMember, error) {
	query := `SELECT id, value, category_id FROM category_members
			  WHERE id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var cm CategoryMember
	err := cmm.db.QueryRowContext(ctx, query, id).Scan(&cm.Id, &cm.Value, &cm.CategoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &cm, nil
}

func (cmm CategoryMemberModel) GetAll() ([]CategoryMember, int, error) {
	query := `SELECT count(*) OVER(), id, value, category_id FROM category_members;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := cmm.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrRecordNotFound
		}
		return nil, 0, err
	}

	var members []CategoryMember
	totalAmount := 0
	for rows.Next() {
		var cm CategoryMember
		err := rows.Scan(
			&totalAmount,
			&cm.Id,
			&cm.Value,
			&cm.CategoryId,
		)
		if err != nil {
			return nil, 0, err
		}
		members = append(members, cm)
	}
	return members, totalAmount, nil
}

func (cmm CategoryMemberModel) Delete(id int64) error {
	query := `DELETE FROM category_members WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := cmm.db.ExecContext(ctx, query, id)
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
