// Package handlers — операции CRUD для таблицы schedule (график сотрудников).
//
// Структура таблицы (см. internal/db/migrations.go):
//
//	schedule(
//	    id        INTEGER PRIMARY KEY AUTOINCREMENT,
//	    user_id   INTEGER NOT NULL REFERENCES staff(id),
//	    date      TEXT    NOT NULL,   -- формат YYYY-MM-DD
//	    shift     TEXT    NOT NULL    -- '1' / 'к' / 'Я' / 'о' / '' и т.п.
//	)
//
// Один день сотрудника — одна строка. Пустая смена = выходной.
package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"OpticERP/internal/db"
	"OpticERP/internal/models"
)

// GetSchedule возвращает все записи графика, опционально фильтруя по
// диапазону дат (в формате YYYY-MM-DD включительно). Если from/to пустые —
// возвращается весь график. Сортировка: сначала user_id, потом date.
func GetSchedule(from, to string) ([]models.Schedule, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, user_id, date, shift FROM schedule`
	args := []any{}
	where := ""
	if from != "" {
		where += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		where += " AND date <= ?"
		args = append(args, to)
	}
	// Убираем ведущий " AND" → заменяем на WHERE при наличии условий.
	if where != "" {
		query += " WHERE " + where[5:]
	}
	query += " ORDER BY user_id, date"

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения графика: %w", err)
	}
	defer rows.Close()

	// Инициализируем пустым слайсом, чтобы в JSON уехал [] (а не null).
	result := make([]models.Schedule, 0)
	for rows.Next() {
		var s models.Schedule
		if err := rows.Scan(&s.ID, &s.UserID, &s.Date, &s.Shift); err != nil {
			return nil, fmt.Errorf("ошибка сканирования графика: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// UpsertSchedule создаёт или обновляет запись графика (user_id, date).
// Если shift == "" — запись удаляется (логика "пустая ячейка = выходной,
// строку не храним"). Возвращает true, если запись фактически сохранена.
func UpsertSchedule(userID int64, date, shift string) (bool, error) {
	log.Printf("[UpsertSchedule] userID=%d date=%q shift=%q", userID, date, shift)
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	if userID <= 0 {
		return false, errors.New("user_id должен быть положительным")
	}
	if date == "" {
		return false, errors.New("дата не может быть пустой")
	}
	// Проверяем, что сотрудник существует (FK на staff).
	var exists int64
	if err := conn.QueryRow(`SELECT COUNT(*) FROM staff WHERE id = ?`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("ошибка проверки сотрудника: %w", err)
	}
	if exists == 0 {
		return false, fmt.Errorf("сотрудник с id=%d не найден", userID)
	}

	// Пустая смена = удаляем строку, чтобы таблица не пухла.
	if shift == "" {
		res, err := conn.Exec(`DELETE FROM schedule WHERE user_id = ? AND date = ?`, userID, date)
		if err != nil {
			return false, fmt.Errorf("ошибка удаления пустой записи графика: %w", err)
		}
		n, _ := res.RowsAffected()
		return n > 0, nil
	}

	// INSERT ... ON CONFLICT — заменяем существующую смену, если пара
	// (user_id, date) уже есть. На старых версиях SQLite (<3.24) ON
	// CONFLICT тоже работает, так что это безопасно.
	_, err = conn.Exec(`
		INSERT INTO schedule (user_id, date, shift)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, date) DO UPDATE SET shift = excluded.shift
	`, userID, date, shift)
	if err != nil {
		// Фоллбэк: возможно, нет уникального индекса (user_id,date).
		// Тогда делаем ручной upsert: пробуем UPDATE, и если ничего не
		// обновилось — INSERT.
		n, uerr := upsertFallback(conn, userID, date, shift)
		if uerr != nil {
			return false, fmt.Errorf("ошибка сохранения графика: %w", uerr)
		}
		return n, nil
	}
	return true, nil
}

// upsertFallback — медленный путь upsert без уникального индекса.
func upsertFallback(conn *sql.DB, userID int64, date, shift string) (bool, error) {
	res, err := conn.Exec(`UPDATE schedule SET shift = ? WHERE user_id = ? AND date = ?`, shift, userID, date)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return true, nil
	}
	if _, err := conn.Exec(`INSERT INTO schedule (user_id, date, shift) VALUES (?, ?, ?)`, userID, date, shift); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteSchedule удаляет запись графика по (user_id, date). Возвращает
// true, если запись существовала и была удалена.
func DeleteSchedule(userID int64, date string) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	res, err := conn.Exec(`DELETE FROM schedule WHERE user_id = ? AND date = ?`, userID, date)
	if err != nil {
		return false, fmt.Errorf("ошибка удаления записи графика: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// DeleteScheduleForUser удаляет весь график конкретного сотрудника.
// Используется при удалении сотрудника из настроек или при очистке графика.
func DeleteScheduleForUser(userID int64) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	res, err := conn.Exec(`DELETE FROM schedule WHERE user_id = ?`, userID)
	if err != nil {
		return false, fmt.Errorf("ошибка очистки графика сотрудника: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
