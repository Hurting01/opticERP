// Package handlers содержит SQL-операции для всех сущностей приложения.
// Каждая функция возвращает ошибку как есть — Wails пробрасывает её
// во фронтенд через Promise-rejection, и JS-код может её показать.
package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"OpticERP/internal/db"
	"OpticERP/internal/models"
)

// GetPositions возвращает все должности.
func GetPositions() ([]models.Position, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`SELECT id, name, created_at, norm_hours, hours_per_shift, salary, additional_payments FROM positions ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения должностей: %w", err)
	}
	defer rows.Close()

	var result []models.Position
	for rows.Next() {
		var p models.Position
		var nh, sal sql.NullInt64
		var hps, ap sql.NullFloat64
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &nh, &hps, &sal, &ap); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		if nh.Valid {
			v := nh.Int64
			p.NormHours = &v
		}
		if hps.Valid {
			v := hps.Float64
			p.HoursPerShift = &v
		}
		if sal.Valid {
			v := sal.Int64
			p.Salary = &v
		}
		if ap.Valid {
			v := ap.Float64
			p.AdditionalPayments = &v
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// CreatePosition создаёт должность.
// Сигнатура 1:1 повторяет Rust-команду create_position:
//   - norm_hours: Option<i32> → *int64
//   - hours_per_shift: Option<f64> → *float64
//   - salary: Option<i32> → *int64
//   - additional_payments: Option<f64> → *float64
func CreatePosition(name string, normHours, salary *int64, hoursPerShift, additionalPayments *float64) (models.Position, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Position{}, err
	}
	if name == "" {
		return models.Position{}, errors.New("название должности не может быть пустым")
	}
	res, err := conn.Exec(
		`INSERT INTO positions (name, norm_hours, hours_per_shift, salary, additional_payments) VALUES (?, ?, ?, ?, ?)`,
		name, nullInt64(normHours), nullFloat64(hoursPerShift), nullInt64(salary), nullFloat64(additionalPayments),
	)
	if err != nil {
		return models.Position{}, fmt.Errorf("ошибка создания должности: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetPositionByID(id)
}

// GetPositionByID возвращает должность по идентификатору.
func GetPositionByID(id int64) (models.Position, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Position{}, err
	}
	row := conn.QueryRow(`SELECT id, name, created_at, norm_hours, hours_per_shift, salary, additional_payments FROM positions WHERE id = ?`, id)
	var p models.Position
	var nh, sal sql.NullInt64
	var hps, ap sql.NullFloat64
	if err := row.Scan(&p.ID, &p.Name, &p.CreatedAt, &nh, &hps, &sal, &ap); err != nil {
		return models.Position{}, fmt.Errorf("должность не найдена: %w", err)
	}
	if nh.Valid {
		v := nh.Int64
		p.NormHours = &v
	}
	if hps.Valid {
		v := hps.Float64
		p.HoursPerShift = &v
	}
	if sal.Valid {
		v := sal.Int64
		p.Salary = &v
	}
	if ap.Valid {
		v := ap.Float64
		p.AdditionalPayments = &v
	}
	return p, nil
}

// UpdatePosition обновляет должность.
func UpdatePosition(positionID int64, positionName string, normHours, salary *int64, hoursPerShift, additionalPayments *float64) (models.Position, error) {
	conn, err := db.DB()
	if err != nil {
		return models.Position{}, err
	}
	_, err = conn.Exec(
		`UPDATE positions SET name = ?, norm_hours = ?, hours_per_shift = ?, salary = ?, additional_payments = ? WHERE id = ?`,
		positionName, nullInt64(normHours), nullFloat64(hoursPerShift), nullInt64(salary), nullFloat64(additionalPayments), positionID,
	)
	if err != nil {
		return models.Position{}, fmt.Errorf("ошибка обновления должности: %w", err)
	}
	return GetPositionByID(positionID)
}

// DeletePosition удаляет должность. Возвращает true при успехе.
func DeletePosition(positionID int64) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	res, err := conn.Exec(`DELETE FROM positions WHERE id = ?`, positionID)
	if err != nil {
		return false, fmt.Errorf("ошибка удаления должности: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// === helpers ===

func nullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullFloat64(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}
