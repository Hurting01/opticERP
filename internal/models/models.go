// Package models содержит структуры сущностей, отражающие таблицы SQLite.
// Имена полей и типы соответствуют оригинальным Rust-моделям из opticTauri
// (src-tauri/src/models.rs) и src-tauri/src/schema.rs.
package models

// === positions ===

// Position — должность. Соответствует таблице positions.
type Position struct {
	ID                 int64    `json:"id"`
	Name               string   `json:"name"`
	CreatedAt          string   `json:"created_at"`
	NormHours          *int64   `json:"norm_hours"`
	HoursPerShift      *float64 `json:"hours_per_shift"`
	Salary             *int64   `json:"salary"`
	AdditionalPayments *float64 `json:"additional_payments"`
}

// === staff ===

// Staff — сотрудник салона оптики.
type Staff struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	PositionID int64  `json:"position_id"`
	IsActive   int64  `json:"is_active"`
	CreatedAt  string `json:"created_at"`
}

// StaffWithPosition — сотрудник с уже подтянутым названием должности.
// Удобно для фронта, чтобы не делать дополнительный запрос positions.
type StaffWithPosition struct {
	ID           int64  `json:"id"`
	FullName     string `json:"full_name"`
	PositionID   int64  `json:"position_id"`
	PositionName string `json:"position_name"`
	IsActive     int64  `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// === tasks (историческое; в текущей версии хранится в localStorage) ===

// Task — задача на день. В Tauri-источнике таблица используется
// как fallback-хранилище, фронт в основном пишет в localStorage.
type Task struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// === sales ===

// Sale — продажа (рецепт, материалы, оплаты).
type Sale struct {
	ID            int64   `json:"id"`
	DateTime      string  `json:"datetime"`
	ProductName   string  `json:"product_name"`
	Recipe        *string `json:"recipe"`
	TotalAmount   float64 `json:"total_amount"`
	AdvanceAmount float64 `json:"advance_amount"`
	CashAmount    float64 `json:"cash_amount"`
	CardAmount    float64 `json:"card_amount"`
	SbpAmount     float64 `json:"sbp_amount"`
	CreatedAt     string  `json:"created_at"`
}

// === daily_workers ===

// DailyWorker — сотрудник, отработавший конкретный день.
type DailyWorker struct {
	ID         int64   `json:"id"`
	Date       string  `json:"date"`
	WorkerName string  `json:"worker_name"`
	Shift      *string `json:"shift"`
}

// === total_income ===

// TotalIncome — дневная выручка (деньги в кассу + безнал).
type TotalIncome struct {
	ID        int64   `json:"id"`
	Date      string  `json:"date"`
	TotalSum  float64 `json:"total_sum"`
	CreatedAt string  `json:"created_at"`
}

// === cash_register ===

// CashRegister — утренняя/вечерняя касса.
type CashRegister struct {
	ID            int64   `json:"id"`
	Date          string  `json:"date"`
	MorningAmount float64 `json:"morning_amount"`
	EveningAmount float64 `json:"evening_amount"`
	CreatedAt     string  `json:"created_at"`
}

// === cash_operations ===

// CashOperation — кассовая операция (приход/расход).
type CashOperation struct {
	ID            int64   `json:"id"`
	Date          string  `json:"date"`
	OperationType string  `json:"operation_type"`
	Amount        float64 `json:"amount"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
}

// === schedule ===

// Schedule — запись расписания сотрудника на дату.
type Schedule struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Date   string `json:"date"`
	Shift  string `json:"shift"`
}

// === bonuses ===

// Bonus — бонус сотрудника за продажу.
type Bonus struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	SaleID    *int64  `json:"sale_id"`
	Amount    float64 `json:"amount"`
	Date      string  `json:"date"`
	CreatedAt string  `json:"created_at"`
}

// === conversion ===

// Conversion — дневная конверсия (посетители, продажи, заказы, диагностики).
type Conversion struct {
	ID                    int64    `json:"id"`
	Date                  string   `json:"date"`
	VisitorsCount         int64    `json:"visitors_count"`
	SalesCount            int64    `json:"sales_count"`
	OrdersCount           int64    `json:"orders_count"`
	DiagnosticsCount      int64    `json:"diagnostics_count"`
	Turnover              float64  `json:"turnover"`
	ConversionVsLastYear  *float64 `json:"conversion_vs_last_year"`
	ConversionVsLastMonth *float64 `json:"conversion_vs_last_month"`
	ConversionVsLastWeek  *float64 `json:"conversion_vs_last_week"`
	CreatedAt             string   `json:"created_at"`
}

// === monthly_plan ===

// MonthlyPlan — месячный план по заказам и обороту.
type MonthlyPlan struct {
	ID                  int64    `json:"id"`
	Year                int64    `json:"year"`
	Month               int64    `json:"month"`
	OrdersPlan          float64  `json:"orders_plan"`
	TurnoverPlan        float64  `json:"turnover_plan"`
	OrdersActual        float64  `json:"orders_actual"`
	TurnoverActual      float64  `json:"turnover_actual"`
	DailyOrdersPlan     *float64 `json:"daily_orders_plan"`
	DailyOrdersActual   float64  `json:"daily_orders_actual"`
	DailyTurnoverPlan   *float64 `json:"daily_turnover_plan"`
	DailyTurnoverActual float64  `json:"daily_turnover_actual"`
	RemainingOrders     *float64 `json:"remaining_orders"`
	RemainingTurnover   *float64 `json:"remaining_turnover"`
	CreatedAt           string   `json:"created_at"`
}

// === weekday_analysis ===

// WeekdayAnalysis — анализ по дням недели.
type WeekdayAnalysis struct {
	ID         int64   `json:"id"`
	Year       int64   `json:"year"`
	Month      int64   `json:"month"`
	Weekday    int64   `json:"weekday"`
	TotalSales float64 `json:"total_sales"`
	OrderCount int64   `json:"order_count"`
	CreatedAt  string  `json:"created_at"`
}

// === conversion_notes ===

// ConversionNote — заметка к дню конверсии.
type ConversionNote struct {
	ID             int64  `json:"id"`
	ConversionDate string `json:"conversion_date"`
	Note           string `json:"note"`
	CreatedAt      string `json:"created_at"`
}

// === salary ===

// Salary — зарплата сотрудника за месяц.
type Salary struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	Month       string  `json:"month"`
	BaseSalary  float64 `json:"base_salary"`
	Bonus       float64 `json:"bonus"`
	Deductions  float64 `json:"deductions"`
	TotalSalary float64 `json:"total_salary"`
	CreatedAt   string  `json:"created_at"`
}

// === realized_positions ===

// RealizedPosition — реализация товара за месяц.
type RealizedPosition struct {
	ID          int64   `json:"id"`
	ProductName string  `json:"product_name"`
	Quantity    int64   `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
	Month       string  `json:"month"`
	CreatedAt   string  `json:"created_at"`
}

// === position_counts ===

// PositionCount — счётчик товара на складе.
type PositionCount struct {
	ID          int64  `json:"id"`
	ProductName string `json:"product_name"`
	Quantity    int64  `json:"quantity"`
	CreatedAt   string `json:"created_at"`
}
