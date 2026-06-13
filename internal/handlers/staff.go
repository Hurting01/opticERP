package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"OpticERP/internal/db"
	"OpticERP/internal/models"
)

// GetStaff возвращает всех сотрудников с подтянутым названием должности.
func GetStaff() ([]models.StaffWithPosition, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`
		SELECT s.id, s.full_name, s.position_id, s.is_active, s.created_at, COALESCE(p.name, '')
		FROM staff s
		LEFT JOIN positions p ON p.id = s.position_id
		ORDER BY s.id`)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения персонала: %w", err)
	}
	defer rows.Close()

	var result []models.StaffWithPosition
	for rows.Next() {
		var s models.StaffWithPosition
		if err := rows.Scan(&s.ID, &s.FullName, &s.PositionID, &s.IsActive, &s.CreatedAt, &s.PositionName); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// CreateStaff добавляет сотрудника.
func CreateStaff(fullName string, positionID int64) (models.StaffWithPosition, error) {
	conn, err := db.DB()
	if err != nil {
		return models.StaffWithPosition{}, err
	}
	if fullName == "" {
		return models.StaffWithPosition{}, errors.New("ФИО сотрудника не может быть пустым")
	}
	res, err := conn.Exec(`INSERT INTO staff (full_name, position_id) VALUES (?, ?)`, fullName, positionID)
	if err != nil {
		return models.StaffWithPosition{}, fmt.Errorf("ошибка создания сотрудника: %w", err)
	}
	id, _ := res.LastInsertId()
	return getStaffByID(id)
}

// UpdateStaff обновляет ФИО и должность сотрудника.
func UpdateStaff(staffID int64, newFullName string, newPositionID int64) (models.StaffWithPosition, error) {
	conn, err := db.DB()
	if err != nil {
		return models.StaffWithPosition{}, err
	}
	_, err = conn.Exec(`UPDATE staff SET full_name = ?, position_id = ? WHERE id = ?`, newFullName, newPositionID, staffID)
	if err != nil {
		return models.StaffWithPosition{}, fmt.Errorf("ошибка обновления сотрудника: %w", err)
	}
	return getStaffByID(staffID)
}

// DeleteStaff удаляет сотрудника.
func DeleteStaff(staffID int64) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	res, err := conn.Exec(`DELETE FROM staff WHERE id = ?`, staffID)
	if err != nil {
		return false, fmt.Errorf("ошибка удаления сотрудника: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func getStaffByID(id int64) (models.StaffWithPosition, error) {
	conn, err := db.DB()
	if err != nil {
		return models.StaffWithPosition{}, err
	}
	row := conn.QueryRow(`
		SELECT s.id, s.full_name, s.position_id, s.is_active, s.created_at, COALESCE(p.name, '')
		FROM staff s LEFT JOIN positions p ON p.id = s.position_id WHERE s.id = ?`, id)
	var s models.StaffWithPosition
	if err := row.Scan(&s.ID, &s.FullName, &s.PositionID, &s.IsActive, &s.CreatedAt, &s.PositionName); err != nil {
		if err == sql.ErrNoRows {
			return models.StaffWithPosition{}, errors.New("сотрудник не найден")
		}
		return models.StaffWithPosition{}, fmt.Errorf("ошибка чтения сотрудника: %w", err)
	}
	return s, nil
}
