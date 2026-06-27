package handlers

import (
	"OpticERP/internal/db"
	"OpticERP/internal/models"
	"database/sql"
	"time"
)

// GetSalesByDate возвращает все продажи за указанную дату (YYYY-MM-DD).
func GetSalesByDate(date string) ([]models.Sale, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`
		SELECT id, datetime, product_name, recipe, total_amount, advance_amount,
		       cash_amount, card_amount, sbp_amount, comment, created_at
		FROM sales
		WHERE DATE(datetime) = ?
		ORDER BY datetime DESC
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var s models.Sale
		if err := rows.Scan(
			&s.ID, &s.DateTime, &s.ProductName, &s.Recipe, &s.TotalAmount,
			&s.AdvanceAmount, &s.CashAmount, &s.CardAmount, &s.SbpAmount, &s.Comment, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}
	return sales, nil
}

// CreateSale создаёт новую продажу.
func CreateSale(
	dateTime string,
	productName string,
	recipe *string,
	totalAmount float64,
	advanceAmount float64,
	cashAmount float64,
	cardAmount float64,
	sbpAmount float64,
	comment *string,
) (models.Sale, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Sale{}, err
	}

	// Если dateTime пуст, используем текущее время
	if dateTime == "" {
		dateTime = time.Now().Format("2006-01-02 15:04:05")
	}

	res, err := conn.Exec(`
		INSERT INTO sales (datetime, product_name, recipe, total_amount, advance_amount,
		                   cash_amount, card_amount, sbp_amount, comment, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, dateTime, productName, recipe, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment)
	if err != nil {
		return models.Sale{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return models.Sale{}, err
	}

	return GetSaleByID(id)
}

// GetSaleByID возвращает продажу по ID.
func GetSaleByID(id int64) (models.Sale, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Sale{}, err
	}
	var s models.Sale
	err = conn.QueryRow(`
		SELECT id, datetime, product_name, recipe, total_amount, advance_amount,
		       cash_amount, card_amount, sbp_amount, comment, created_at
		FROM sales
		WHERE id = ?
	`, id).Scan(
		&s.ID, &s.DateTime, &s.ProductName, &s.Recipe, &s.TotalAmount,
		&s.AdvanceAmount, &s.CashAmount, &s.CardAmount, &s.SbpAmount, &s.Comment, &s.CreatedAt,
	)
	if err != nil {
		return models.Sale{}, err
	}
	return s, nil
}

// UpdateSale обновляет существующую продажу.
func UpdateSale(
	id int64,
	dateTime string,
	productName string,
	recipe *string,
	totalAmount float64,
	advanceAmount float64,
	cashAmount float64,
	cardAmount float64,
	sbpAmount float64,
	comment *string,
) (models.Sale, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Sale{}, err
	}

	_, err = conn.Exec(`
		UPDATE sales
		SET datetime = ?, product_name = ?, recipe = ?, total_amount = ?,
		    advance_amount = ?, cash_amount = ?, card_amount = ?, sbp_amount = ?, comment = ?
		WHERE id = ?
	`, dateTime, productName, recipe, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment, id)
	if err != nil {
		return models.Sale{}, err
	}

	return GetSaleByID(id)
}

// DeleteSale удаляет продажу по ID.
func DeleteSale(id int64) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	res, err := conn.Exec("DELETE FROM sales WHERE id = ?", id)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// GetSalesByDateRange возвращает продажи в диапазоне дат.
func GetSalesByDateRange(from string, to string) ([]models.Sale, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}
	
	query := `
		SELECT id, datetime, product_name, recipe, total_amount, advance_amount,
		       cash_amount, card_amount, sbp_amount, comment, created_at
		FROM sales
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	if from != "" {
		query += " AND DATE(datetime) >= ?"
		args = append(args, from)
	}
	
	if to != "" {
		query += " AND DATE(datetime) <= ?"
		args = append(args, to)
	}
	
	query += " ORDER BY datetime DESC"
	
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var s models.Sale
		var recipe sql.NullString
		var comment sql.NullString
		
		if err := rows.Scan(
			&s.ID, &s.DateTime, &s.ProductName, &recipe, &s.TotalAmount,
			&s.AdvanceAmount, &s.CashAmount, &s.CardAmount, &s.SbpAmount, &comment, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		
		if recipe.Valid {
			s.Recipe = &recipe.String
		}
		
		if comment.Valid {
			s.Comment = &comment.String
		}
		
		sales = append(sales, s)
	}
	
	return sales, nil
}
