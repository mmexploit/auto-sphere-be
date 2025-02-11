package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Mahider-T/autoSphere/validator"
	"github.com/lib/pq"
)

type ShopApprovalStatus string

const (
	APPROVED ShopApprovalStatus = "APPROVED"
	PENDING  ShopApprovalStatus = "PENDING"
	DECLINED ShopApprovalStatus = "DECLINED"
)

type Shop struct {
	Id              int                `json:"id"`
	Name            string             `json:"name"`
	Phone_Number    string             `json:"phone_number"`
	Email           string             `json:"email"`
	Approval_Status ShopApprovalStatus `json:"approval_status"`
	Location        string             `json:"location"`
	Coordinate      string             `json:"coordinate"` // Should be in "longitude latitude" format
	Thumbnail       *string            `json:"thumbnail"`
	Photos          []string           `json:"photos"`
	Created_At      time.Time          `json:"-"`
	Created_By      int                `json:"created_by"`
}

func ValidateShop(v *validator.Validator, sh *Shop) {
	var phoneRegex = regexp.MustCompile(`^(09|07)\d{8}$`)
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	var coordinateRegex = regexp.MustCompile(`^\d+ \d+$`)

	v.Check(sh.Name != "", "name", "name must not be empty")
	v.Check(sh.Location != "", "location", "location must not be empty")
	v.Check(sh.Coordinate != "", "coordinate", "coordinate can not me empty")
	v.Check(validator.Matches(sh.Coordinate, coordinateRegex), "coordinate", "coordinate must be logitude and latitute separated by space")
	v.Check(validator.Matches(sh.Phone_Number, phoneRegex), "phone_number", "phone number must start with 07 or 09 and must be 10 digits long")
	v.Check(validator.Matches(sh.Email, emailRegex), "email", "not a valid email")
}

type ShopModel struct {
	db *sql.DB
}

func (sh ShopModel) Create(shop *Shop) error {
	query := `
		INSERT INTO shops 
			(name, phone_number, email, location, coordinate, thumbnail,photos, approval_status, created_by)
		VALUES 
			($1, $2, $3, $4, ST_GeogFromText($5), $6, $7, $8, $9)
		RETURNING id, name, phone_number, email, location, ST_AsText(coordinate), thumbnail, photos, created_at, approval_status, created_by;`

	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()
	args := []interface{}{
		shop.Name,
		shop.Phone_Number,
		shop.Email,
		shop.Location,
		fmt.Sprintf("SRID=4326;POINT(%s)", shop.Coordinate),
		shop.Thumbnail,
		pq.Array(shop.Photos),
		shop.Approval_Status,
		shop.Created_By,
	}
	return sh.db.QueryRowContext(ctx, query, args...).Scan(&shop.Id, &shop.Name, &shop.Phone_Number, &shop.Email, &shop.Location, &shop.Coordinate, &shop.Thumbnail, pq.Array(&shop.Photos), &shop.Created_At, &shop.Approval_Status, &shop.Created_By)
}

func (sh ShopModel) Get(id int64) (*Shop, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, name, phone_number, email, location, ST_AsText(coordinate), thumbnail, photos, created_at, approval_status, created_by FROM shops WHERE id=$1`

	var shop Shop
	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()

	err := sh.db.QueryRowContext(ctx, query, id).Scan(
		&shop.Id, &shop.Name, &shop.Phone_Number, &shop.Email,
		&shop.Location, &shop.Coordinate, &shop.Thumbnail, pq.Array(&shop.Photos), &shop.Created_At, &shop.Approval_Status, &shop.Created_By,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &shop, nil
}
func (sh ShopModel) Patch(shop *Shop) error {
	query := `
		UPDATE shops 
		SET name=$1, phone_number=$2, email=$3, location=$4, coordinate=ST_GeogFromText($5), thumbnail=$6, photos=$7
		WHERE id=$8
		RETURNING id, name, phone_number, email, location, ST_AsText(coordinate), thumbnail, photos, created_at, approval_status, created_by;
	`

	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()

	args := []interface{}{
		shop.Name,
		shop.Phone_Number,
		shop.Email,
		shop.Location,
		fmt.Sprintf("SRID=4326;POINT(%s)", shop.Coordinate),
		shop.Thumbnail,
		pq.Array(shop.Photos),
		shop.Id,
	}

	return sh.db.QueryRowContext(ctx, query, args...).Scan(
		&shop.Id, &shop.Name, &shop.Phone_Number, &shop.Email,
		&shop.Location, &shop.Coordinate, &shop.Thumbnail, pq.Array(&shop.Photos), &shop.Created_At, &shop.Approval_Status, &shop.Created_By,
	)
}
func (sh ShopModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()

	query := `DELETE FROM shops WHERE id=$1`
	result, err := sh.db.ExecContext(ctx, query, id)
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

func (sh ShopModel) GetAll(name, coordinate string, maxDistance int, filters Filters, categoryValues []string) ([]Shop, Metadata, error) {
	fmt.Print("Category members are :- ", categoryValues)
	var query string
	var args []interface{}

	// baseQuery with fully qualified references for ids
	baseQuery := `SELECT count(*) OVER (), shops.id AS shop_id, shops.name, shops.phone_number, shops.email, shops.location, 
				  ST_AsText(shops.coordinate), shops.thumbnail, shops.photos, shops.created_at, shops.approval_status, shops.created_by
				  FROM shops
				  LEFT JOIN shop_categories ON shops.id = shop_categories.shop_id
				  LEFT JOIN category_members ON shop_categories.category_member_id = category_members.id
				  WHERE (to_tsvector('simple', shops.name) @@ plainto_tsquery('simple', $1) OR $1 = '')`
	args = append(args, name)

	if len(categoryValues) > 0 {
		fmt.Print("HELOOOOO BRUV")
		baseQuery += ` AND category_members.value IN (` + generatePlaceholders(len(categoryValues)) + `)`
		for _, value := range categoryValues {
			args = append(args, value)
		}

		// Group by shop ID and ensure the shop has all the required category values
		baseQuery += ` GROUP BY shops.id HAVING COUNT(DISTINCT category_members.value) = ` + fmt.Sprintf("%d", len(categoryValues))
	}

	if coordinate != "" {
		// Add the maxDistance filter to the query
		query = fmt.Sprintf(`%s AND ST_Distance(shops.coordinate, ST_GeogFromText($%d)) <= $%d
							ORDER BY ST_Distance(shops.coordinate, ST_GeogFromText($%d)) ASC, shops.id ASC
							LIMIT $%d OFFSET $%d`,
			baseQuery, len(args)+1, len(args)+2, len(args)+1, len(args)+3, len(args)+4)

		// Add the maxDistance argument and the rest of the pagination arguments
		args = append(args, fmt.Sprintf("SRID=4326;POINT(%s)", coordinate), maxDistance, filters.limit(), filters.offset())
	} else {
		query = fmt.Sprintf(`%s ORDER BY shops.id ASC
							LIMIT $%d OFFSET $%d`,
			baseQuery, len(args)+1, len(args)+2)

		args = append(args, filters.limit(), filters.offset())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Printf("Base query is here bro \n %s\n\n", baseQuery)

	rows, err := sh.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	shops := []Shop{}
	for rows.Next() {
		var shop Shop
		err := rows.Scan(
			&totalRecords,
			&shop.Id, // This corresponds to shops.id as shop_id
			&shop.Name,
			&shop.Phone_Number,
			&shop.Email,
			&shop.Location,
			&shop.Coordinate,
			&shop.Thumbnail,
			pq.Array(&shop.Photos),
			&shop.Created_At,
			&shop.Approval_Status,
			&shop.Created_By,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		shops = append(shops, shop)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := filters.calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return shops, metadata, nil
}

func (sh ShopModel) UpdateAppoval(id int64, approvalStatus ShopApprovalStatus) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := sh.Get(id)
	if err != nil {
		return err
	}

	query := `UPDATE shops 
			  SET approval_status=$1 RETURNING id, approval_status`

	return sh.db.QueryRowContext(ctx, query, approvalStatus).Scan(&id, &approvalStatus)
}

func generatePlaceholders(n int) string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+2) // Start from $2 to avoid conflicting with existing $1
	}
	return strings.Join(placeholders, ", ")
}
