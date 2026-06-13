package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// Migrate создаёт и обновляет все таблицы БД.
// Это прямой порт функции run_migrations() из opticTauri/src-tauri/src/lib.rs:
// 1:1 воспроизведена логика миграций (с условными ветками для старых
// версий таблицы positions) и защитная очистка данных.
func Migrate(conn *sql.DB) error {
	stmts := []string{
		// tasks
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			description TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES staff(id)
		)`,
		// sales
		`CREATE TABLE IF NOT EXISTS sales (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			datetime TEXT NOT NULL,
			product_name TEXT NOT NULL,
			recipe TEXT,
			total_amount REAL NOT NULL,
			advance_amount REAL NOT NULL DEFAULT 0,
			cash_amount REAL NOT NULL DEFAULT 0,
			card_amount REAL NOT NULL DEFAULT 0,
			sbp_amount REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// daily_workers
		`CREATE TABLE IF NOT EXISTS daily_workers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			worker_name TEXT NOT NULL,
			shift TEXT
		)`,
		// total_income
		`CREATE TABLE IF NOT EXISTS total_income (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			total_sum REAL NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// cash_register
		`CREATE TABLE IF NOT EXISTS cash_register (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			morning_amount REAL NOT NULL,
			evening_amount REAL NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// cash_operations
		`CREATE TABLE IF NOT EXISTS cash_operations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			operation_type TEXT NOT NULL,
			amount REAL NOT NULL,
			description TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// schedule
		`CREATE TABLE IF NOT EXISTS schedule (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			shift TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES staff(id)
		)`,
		// bonuses
		`CREATE TABLE IF NOT EXISTS bonuses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			sale_id INTEGER,
			amount REAL NOT NULL,
			date TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (user_id) REFERENCES staff(id),
			FOREIGN KEY (sale_id) REFERENCES sales(id)
		)`,
		// conversion
		`CREATE TABLE IF NOT EXISTS conversion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL UNIQUE,
			visitors_count INTEGER NOT NULL DEFAULT 0,
			sales_count INTEGER NOT NULL DEFAULT 0,
			orders_count INTEGER NOT NULL DEFAULT 0,
			diagnostics_count INTEGER NOT NULL DEFAULT 0,
			turnover REAL NOT NULL DEFAULT 0,
			conversion_vs_last_year REAL,
			conversion_vs_last_month REAL,
			conversion_vs_last_week REAL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// monthly_plan
		`CREATE TABLE IF NOT EXISTS monthly_plan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			orders_plan REAL NOT NULL,
			turnover_plan REAL NOT NULL,
			orders_actual REAL NOT NULL DEFAULT 0,
			turnover_actual REAL NOT NULL DEFAULT 0,
			daily_orders_plan REAL,
			daily_orders_actual REAL DEFAULT 0,
			daily_turnover_plan REAL,
			daily_turnover_actual REAL DEFAULT 0,
			remaining_orders REAL,
			remaining_turnover REAL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// weekday_analysis
		`CREATE TABLE IF NOT EXISTS weekday_analysis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			weekday INTEGER NOT NULL,
			total_sales REAL NOT NULL DEFAULT 0,
			order_count INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// conversion_notes
		`CREATE TABLE IF NOT EXISTS conversion_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversion_date TEXT NOT NULL,
			note TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (conversion_date) REFERENCES conversion(date)
		)`,
		// salary
		`CREATE TABLE IF NOT EXISTS salary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			month TEXT NOT NULL,
			base_salary REAL NOT NULL,
			bonus REAL NOT NULL DEFAULT 0,
			deductions REAL NOT NULL DEFAULT 0,
			total_salary REAL NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (user_id) REFERENCES staff(id)
		)`,
		// realized_positions
		`CREATE TABLE IF NOT EXISTS realized_positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			total_amount REAL NOT NULL,
			month TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
		// position_counts
		`CREATE TABLE IF NOT EXISTS position_counts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_name TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)`,
	}

	for _, s := range stmts {
		if _, err := conn.Exec(s); err != nil {
			return fmt.Errorf("ошибка миграции: %w\nSQL: %s", err, s)
		}
	}

	// Сложная ветка миграции positions (1:1 из lib.rs).
	if err := migratePositions(conn); err != nil {
		return err
	}

	// Удаляем устаревшую таблицу users, если она осталась от предыдущих версий.
	if _, err := conn.Exec(`DROP TABLE IF EXISTS users`); err != nil {
		return fmt.Errorf("ошибка удаления таблицы users: %w", err)
	}

	// Таблица staff (последняя, т.к. ссылается на positions).
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS staff (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		position_id INTEGER NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
		FOREIGN KEY (position_id) REFERENCES positions(id)
	)`); err != nil {
		return fmt.Errorf("ошибка создания таблицы staff: %w", err)
	}

	return nil
}

// hasColumn провереряет наличие колонки columnName в таблице tableName.
// Аналог одноимённой функции из lib.rs (через pragma_table_info).
func hasColumn(conn *sql.DB, tableName, columnName string) (bool, error) {
	safeTable := strings.ReplaceAll(tableName, `"`, `""`)
	safeCol := strings.ReplaceAll(columnName, `"`, `""`)
	sql := fmt.Sprintf(
		"SELECT COUNT(*) AS cnt FROM pragma_table_info(\"%s\") WHERE name = '%s'",
		safeTable, safeCol,
	)
	var cnt int64
	if err := conn.QueryRow(sql).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// tableExists проверяет наличие таблицы в sqlite_master.
func tableExists(conn *sql.DB, name string) (bool, error) {
	var cnt int64
	err := conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?",
		name,
	).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// migratePositions повторяет логику большой миграции таблицы positions
// из opticTauri: проверяет наличие старых колонок (Hour_norm, sallary,
// sallary bonus, sallary_bonus, norm_hours_consultant/optometrist и т.д.)
// и при необходимости пересоздаёт таблицу, подставляя выражения
// COALESCE(...) для сохранения данных.
func migratePositions(conn *sql.DB) error {
	exists, err := tableExists(conn, "positions")
	if err != nil {
		return err
	}

	if !exists {
		// Создаём таблицу с актуальной структурой.
		_, err := conn.Exec(`CREATE TABLE positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
			norm_hours INTEGER,
			hours_per_shift REAL DEFAULT 12,
			salary INTEGER,
			additional_payments REAL DEFAULT 5000
		)`)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы positions: %w", err)
		}
		return nil
	}

	// Проверяем наличие старых колонок через PRAGMA (безопасный способ,
	// в отличие от слепого SELECT, который может вернуть Ok на некоторых
	// версиях SQLite и привести к ложному срабатыванию миграции).
	hasHourNorm, _ := hasColumn(conn, "positions", "Hour_norm")
	hasSallary, _ := hasColumn(conn, "positions", "sallary")
	hasSallaryBonusSpace, _ := hasColumn(conn, "positions", "sallary bonus")
	hasSallaryBonus, _ := hasColumn(conn, "positions", "sallary_bonus")
	hasNormHoursConsultant, _ := hasColumn(conn, "positions", "norm_hours_consultant")
	hasNormHoursOptometrist, _ := hasColumn(conn, "positions", "norm_hours_optometrist")
	hasNormHours, _ := hasColumn(conn, "positions", "norm_hours")
	hasSalaryConsultant, _ := hasColumn(conn, "positions", "salary_consultant")
	hasSalaryOptometrist, _ := hasColumn(conn, "positions", "salary_optometrist")
	hasSalary, _ := hasColumn(conn, "positions", "salary")
	hasAdditionalPayments, _ := hasColumn(conn, "positions", "additional_payments")
	hasManagerBonus, _ := hasColumn(conn, "positions", "manager_bonus")

	hasAnyOld := hasHourNorm || hasSallary || hasSallaryBonusSpace || hasSallaryBonus ||
		hasNormHoursConsultant || hasNormHoursOptometrist

	if hasAnyOld {
		// Полная миграция: пересоздаём таблицу с актуальной структурой.
		if _, err := conn.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
			return fmt.Errorf("ошибка отключения FK: %w", err)
		}
		// Удаляем временные таблицы, если они остались с прошлой попытки.
		_, _ = conn.Exec(`DROP TABLE IF EXISTS positions_new`)

		if _, err := conn.Exec(`CREATE TABLE positions_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
			norm_hours INTEGER,
			hours_per_shift REAL DEFAULT 12,
			salary INTEGER,
			additional_payments REAL DEFAULT 5000
		)`); err != nil {
			return fmt.Errorf("ошибка создания positions_new: %w", err)
		}

		// Строим выражения для копирования в новые колонки.
		normHoursExpr := "NULL"
		switch {
		case hasNormHoursConsultant:
			normHoursExpr = "COALESCE(norm_hours_consultant, norm_hours_optometrist)"
		case hasHourNorm:
			normHoursExpr = `"Hour_norm"`
		}

		salaryExpr := "0"
		switch {
		case hasSalaryConsultant && hasSalaryOptometrist:
			salaryExpr = "COALESCE(MAX(salary_consultant, salary_optometrist), 0)"
		case hasSalaryConsultant:
			salaryExpr = "COALESCE(salary_consultant, 0)"
		case hasSalaryOptometrist:
			salaryExpr = "COALESCE(salary_optometrist, 0)"
		case hasSallary:
			salaryExpr = "COALESCE(sallary, 0)"
		}

		additionalExpr := "5000"
		switch {
		case hasSallaryBonusSpace:
			additionalExpr = `"sallary bonus"`
		case hasSallaryBonus:
			additionalExpr = "sallary_bonus"
		case hasManagerBonus:
			additionalExpr = "manager_bonus"
		}

		copySQL := fmt.Sprintf(`INSERT INTO positions_new
			(id, name, created_at, norm_hours, hours_per_shift, salary, additional_payments)
			SELECT id, name, created_at, %s, hours_per_shift, %s, %s
			FROM positions`, normHoursExpr, salaryExpr, additionalExpr)
		if _, err := conn.Exec(copySQL); err != nil {
			return fmt.Errorf("ошибка копирования данных positions: %w", err)
		}
		if _, err := conn.Exec(`DROP TABLE positions`); err != nil {
			return fmt.Errorf("ошибка удаления старой positions: %w", err)
		}
		if _, err := conn.Exec(`ALTER TABLE positions_new RENAME TO positions`); err != nil {
			return fmt.Errorf("ошибка переименования positions_new: %w", err)
		}
		if _, err := conn.Exec(`PRAGMA foreign_keys = ON`); err != nil {
			return fmt.Errorf("ошибка включения FK: %w", err)
		}
	} else {
		// Таблица уже в новой структуре — убеждаемся, что все нужные колонки есть.
		if !hasNormHours {
			_, _ = conn.Exec(`ALTER TABLE positions ADD COLUMN norm_hours INTEGER`)
		}
		if !hasSalary {
			if _, err := conn.Exec(`ALTER TABLE positions ADD COLUMN salary INTEGER`); err != nil {
				return fmt.Errorf("ошибка добавления столбца salary: %w", err)
			}
		}
		// Если остались старые salary_consultant/optometrist — мигрируем данные.
		if hasSalaryConsultant || hasSalaryOptometrist {
			var migrateSQL string
			switch {
			case hasSalaryConsultant && hasSalaryOptometrist:
				migrateSQL = "UPDATE positions SET salary = COALESCE(MAX(salary_consultant, salary_optometrist), 0) WHERE salary IS NULL OR typeof(salary) = 'text'"
			case hasSalaryConsultant:
				migrateSQL = "UPDATE positions SET salary = COALESCE(salary_consultant, 0) WHERE salary IS NULL OR typeof(salary) = 'text'"
			default:
				migrateSQL = "UPDATE positions SET salary = COALESCE(salary_optometrist, 0) WHERE salary IS NULL OR typeof(salary) = 'text'"
			}
			_, _ = conn.Exec(migrateSQL)
		}
		if !hasAdditionalPayments {
			_, _ = conn.Exec(`ALTER TABLE positions ADD COLUMN additional_payments REAL DEFAULT 5000`)
		}
		// Если остался manager_bonus — переносим в additional_payments и удаляем колонку.
		if hasManagerBonus {
			_, _ = conn.Exec(`UPDATE positions SET additional_payments = manager_bonus WHERE additional_payments IS NULL AND manager_bonus IS NOT NULL`)

			// SQLite < 3.35 не поддерживает DROP COLUMN, поэтому пересоздаём.
			if exists, _ := hasColumn(conn, "positions", "id"); exists {
				_, _ = conn.Exec(`PRAGMA foreign_keys = OFF`)
				_, _ = conn.Exec(`DROP TABLE IF EXISTS positions_migrate`)
				_, _ = conn.Exec(`CREATE TABLE positions_migrate (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL UNIQUE,
					created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
					norm_hours INTEGER,
					hours_per_shift REAL DEFAULT 12,
					salary INTEGER,
					additional_payments REAL DEFAULT 5000
				)`)
				_, _ = conn.Exec(`INSERT INTO positions_migrate
					(id, name, created_at, norm_hours, hours_per_shift, salary, additional_payments)
					SELECT id, name, created_at, norm_hours, hours_per_shift,
						COALESCE(salary, 0),
						COALESCE(additional_payments, 5000)
					FROM positions`)
				_, _ = conn.Exec(`DROP TABLE positions`)
				_, _ = conn.Exec(`ALTER TABLE positions_migrate RENAME TO positions`)
				_, _ = conn.Exec(`PRAGMA foreign_keys = ON`)
			}
		}
	}

	// Защитная очистка данных: после неудачной миграции значения
	// могут оказаться нечисловыми. Приводим такие значения к дефолтам.
	_, _ = conn.Exec(`UPDATE positions SET additional_payments = 5000 WHERE typeof(additional_payments) = 'text' OR additional_payments IS NULL`)
	_, _ = conn.Exec(`UPDATE positions SET salary = COALESCE(salary, 0) WHERE typeof(salary) = 'text' OR salary IS NULL`)

	return nil
}
