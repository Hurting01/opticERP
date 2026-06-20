// Package handlers — операции CRUD для таблицы schedule (график сотрудников).
//
// Структура таблицы (см. internal/db/migrations.go):
//
//	schedule(
//	    id        INTEGER PRIMARY KEY AUTOINCREMENT,
//	    user_id   INTEGER NOT NULL REFERENCES staff(id),
//	    date      TEXT    NOT NULL,   -- формат YYYY-MM-DD
//	    shift     TEXT    NOT NULL    -- '1' / 'к' / 'Я' / 'о' / '' и т.п.
//	    hours     REAL    DEFAULT 0,  -- часы в смену, берутся из positions.hours_per_shift
//	    is_working_day INTEGER DEFAULT 1,
//	)
//
// Один день сотрудника — одна строка. Пустая смена = выходной.
//
// Часы (hours) записи графика хранятся в schedule, но фактически
// копируютс�� из positions.hours_per_shift — должности сотрудника.
// При изменении должности сотрудника или hours_per_shift у должности
// все связанные записи schedule автоматически пересчитываются (см.
// recalculateScheduleHoursForUser и RecalculateScheduleHoursForPosition).
package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"OpticERP/internal/db"
	"OpticERP/internal/models"
)

// defaultHoursPerShift — дефолт для hours_per_shift, если в positions
// значение не задано. Совпадает с дефолтом в миграции positions (DEFAULT 12).
const defaultHoursPerShift = 12.0

// parseScheduleMonth вытаскивает год и месяц из даты YYYY-MM-DD.
func parseScheduleMonth(date string) (int64, int64, error) {
	parts := strings.Split(date, "-")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("некорректная дата графика: %q", date)
	}
	year, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("некорректный год графика: %q", date)
	}
	month, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("некорректный месяц графика: %q", date)
	}
	return year, month, nil
}

// resolvePositionHours возвращает hours_per_shift для должности сотрудника.
// Если у сотрудника нет должности или hours_per_shift = NULL, возвращает
// дефолт defaultHoursPerShift. Это значение всегда отражает текущее
// состояние positions — при смене должности сотрудника или редактировании
// hours_per_shift записи в schedule автоматически подтянут новый норматив.
func resolvePositionHours(conn *sql.DB, userID int64) (float64, error) {
	var hps sql.NullFloat64
	err := conn.QueryRow(`
		SELECT p.hours_per_shift
		FROM staff s
		LEFT JOIN positions p ON p.id = s.position_id
		WHERE s.id = ?
	`, userID).Scan(&hps)
	if err != nil {
		if err == sql.ErrNoRows {
			return defaultHoursPerShift, fmt.Errorf("сотрудник с id=%d не найден", userID)
		}
		return defaultHoursPerShift, fmt.Errorf("ош��бка чтения hours_per_shift: %w", err)
	}
	if !hps.Valid || hps.Float64 <= 0 {
		return defaultHoursPerShift, nil
	}
	return hps.Float64, nil
}

// recalculateScheduleHoursForUser пересчитывает hours для всех рабочих дней
// (is_working_day = 1) конкретного сотрудника в schedule, подтягивая
// актуальный hours_per_shift из positions. Используется после смены должности
// сотрудника (UpdateStaff) или редактирования hours_per_shift у должности.
func recalculateScheduleHoursForUser(conn *sql.DB, userID int64) error {
	hours, err := resolvePositionHours(conn, userID)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		UPDATE schedule
		SET hours = ?
		WHERE user_id = ? AND is_working_day = 1
	`, hours, userID)
	if err != nil {
		return fmt.Errorf("ошибка пересчёта часов графика сотрудника: %w", err)
	}
	return nil
}

// RecalculateScheduleHoursForPosition пересчитывает hours во всех записях
// schedule всех сотрудников указанной должности. Вызывается после
// UpdatePosition, чтобы при изменении hours_per_shift у должности все
// связанные записи графика автоматически получили новое значение.
func RecalculateScheduleHoursForPosition(conn *sql.DB, positionID int64) error {
	var hps sql.NullFloat64
	err := conn.QueryRow(`SELECT hours_per_shift FROM positions WHERE id = ?`, positionID).Scan(&hps)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("ошибка чтения hours_per_shift: %w", err)
	}
	hours := defaultHoursPerShift
	if hps.Valid && hps.Float64 > 0 {
		hours = hps.Float64
	}
	_, err = conn.Exec(`
		UPDATE schedule
		SET hours = ?
		WHERE user_id IN (SELECT id FROM staff WHERE position_id = ?)
		  AND is_working_day = 1
	`, hours, positionID)
	if err != nil {
		return fmt.Errorf("ошибка пересчёта часов графика по должности: %w", err)
	}
	return nil
}

// RecalculateScheduleHoursForUser экспортирует пересчёт часов для
// сотрудника наружу (используется после UpdateStaff).
func RecalculateScheduleHoursForUser(userID int64) error {
	conn, err := db.DB()
	if err != nil {
		return err
	}
	return recalculateScheduleHoursForUser(conn, userID)
}

func staffExists(conn *sql.DB, userID int64) error {
	if userID <= 0 {
		return errors.New("user_id должен быть положительным")
	}
	var exists int64
	if err := conn.QueryRow(`SELECT COUNT(*) FROM staff WHERE id = ?`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("ошибка проверки сотрудника: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("сотрудник с id=%d не найден", userID)
	}
	return nil
}

// GetSchedule возвращает все записи графика, опционально фильтруя по
// диапазону дат (в формате YYYY-MM-DD включительно). Если from/to пустые —
// возвращается весь график. Сортировка: сначала user_id, потом date.
//
// Значение hours в каждой записи — это снимок на момент сохранения. Если
// должность сотрудника или hours_per_shift у должности поменялись, можно
// пересчитать часы через RecalculateScheduleHoursForPosition/User.
func GetSchedule(from, to string) ([]models.Schedule, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, user_id, date, shift, hours, is_working_day FROM schedule`
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
		if err := rows.Scan(&s.ID, &s.UserID, &s.Date, &s.Shift, &s.Hours, &s.IsWorkingDay); err != nil {
			return nil, fmt.Errorf("ошибка сканирования графика: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetScheduleMembers возвращает id сотрудников, явно добавленных в график месяца.
func GetScheduleMembers(year, month int64) ([]int64, error) {
	conn, err := db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(`SELECT user_id FROM schedule_members WHERE year = ? AND month = ? ORDER BY user_id`, year, month)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения сотрудников графика: %w", err)
	}
	defer rows.Close()

	result := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("ошибка сканирования сотрудника графика: %w", err)
		}
		result = append(result, userID)
	}
	return result, rows.Err()
}

// AddScheduleMember сохраняет факт добавления сотрудника в график месяца.
func AddScheduleMember(userID, year, month int64) (bool, error) {
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	if err := staffExists(conn, userID); err != nil {
		return false, err
	}
	if year <= 0 || month < 1 || month > 12 {
		return false, errors.New("некорректный месяц графика")
	}
	_, err = conn.Exec(`
		INSERT INTO schedule_members (user_id, year, month)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, year, month) DO NOTHING
	`, userID, year, month)
	if err != nil {
		return false, fmt.Errorf("ошибка добавления сотрудника в график: %w", err)
	}
	return true, nil
}

// UpsertSchedule создаёт или обновляет запись графика (user_id, date).
// Если shift == "" — запись удаляется (логика "пустая ячейка = выходной,
// строку не храним").
//
// ВАЖНО: параметр hours игнорируется — значение часов всегда берётся из
// positions.hours_per_shift текущей должности сотрудника. Это гарантиру��т,
// что в schedule всегда актуальный норматив часов за смену.
//
// Возвращает true, если запись фактически сохранена.
func UpsertSchedule(userID int64, date, shift string, hours float64, isWorkingDay bool) (bool, error) {
	log.Printf("[UpsertSchedule] userID=%d date=%q shift=%q hours=%f is_working_day=%t", userID, date, shift, hours, isWorkingDay)
	conn, err := db.DB()
	if err != nil {
		return false, err
	}
	if date == "" {
		return false, errors.New("дата не может быть пустой")
	}
	if err := staffExists(conn, userID); err != nil {
		return false, err
	}
	if year, month, err := parseScheduleMonth(date); err == nil {
		if _, err := AddScheduleMember(userID, year, month); err != nil {
			return false, err
		}
	}

	// Пустая смена = удаляем строку, чтобы таблица не пухла.
	if shift == "" {
		res, err := conn.Exec(`DELETE FROM schedule WHERE user_id = ? AND date = ?`, userID, date)
		if err != nil {
			return false, fmt.Errorf("ошибка уда��ения пустой записи графика: %w", err)
		}
		n, _ := res.RowsAffected()
		return n > 0, nil
	}

	// Часы берём из текущей должности сотрудника, а не из переданного
	// параметра. Если у сотрудника нет должности / hours_per_shift не задан —
	// используется дефолт defaultHoursPerShift.
	hoursPerShift, herr := resolvePositionHours(conn, userID)
	if herr != nil {
		// staffExists уже отработал выше, но если что-то пошло не так —
		// возвращаем ошибку без записи.
		return false, herr
	}

	// Преобразуем is_working_day в INTEGER
	var isWorkingDayInt int
	if isWorkingDay {
		isWorkingDayInt = 1
	}

	// INSERT ... ON CONFLICT — заменяем существующую смену, если пара
	// (user_id, date) уже есть.
	_, err = conn.Exec(`
		INSERT INTO schedule (user_id, date, shift, hours, is_working_day)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, date) DO UPDATE SET
			shift = excluded.shift,
			hours = excluded.hours,
			is_working_day = excluded.is_working_day
	`, userID, date, shift, hoursPerShift, isWorkingDayInt)
	if err != nil {
		// Фоллбэк
		n, uerr := upsertFallback(conn, userID, date, shift, hoursPerShift)
		if uerr != nil {
			return false, fmt.Errorf("ошибка сохранения графика: %w", uerr)
		}
		return n, nil
	}
	return true, nil
}

// upsertFallback — медленный путь upsert без уникального индекса.
func upsertFallback(conn *sql.DB, userID int64, date, shift string, hours float64) (bool, error) {
	res, err := conn.Exec(`UPDATE schedule SET shift = ?, hours = ?, is_working_day = 0 WHERE user_id = ? AND date = ?`, shift, hours, userID, date)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return true, nil
	}
	if _, err := conn.Exec(`INSERT INTO schedule (user_id, date, shift, hours, is_working_day) VALUES (?, ?, ?, ?, 1)`, userID, date, shift, hours); err != nil {
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
	if _, err := conn.Exec(`DELETE FROM schedule_members WHERE user_id = ?`, userID); err != nil {
		return false, fmt.Errorf("ошибка удаления сотрудника из графика: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
