package main

import (
	"context"
	"fmt"

	"OpticERP/internal/handlers"
	"OpticERP/internal/models"
)

// App — корневая структура Wails-приложения.
// Все экспортируемые методы автоматически становятся доступны
// во фронтенде через сгенерированный биндинг.
type App struct {
	ctx context.Context
}

// NewApp создаёт новый экземпляр приложения.
func NewApp() *App {
	return &App{}
}

// startup вызывается Wails при инициализации WebView2. Здесь сохраняем
// контекст, который позже можно использовать для runtime.LogInfo и т.п.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// === Должности ===

// GetPositions — список должностей (экспорт Wails).
func (a *App) GetPositions() ([]models.Position, error) {
	return handlers.GetPositions()
}

// CreatePosition — создать должность.
func (a *App) CreatePosition(name string, normHours, salary *int64, hoursPerShift, additionalPayments *float64) (models.Position, error) {
	return handlers.CreatePosition(name, normHours, salary, hoursPerShift, additionalPayments)
}

// UpdatePosition — обновить должность.
func (a *App) UpdatePosition(positionId int64, positionName string, normHours, salary *int64, hoursPerShift, additionalPayments *float64) (models.Position, error) {
	return handlers.UpdatePosition(positionId, positionName, normHours, salary, hoursPerShift, additionalPayments)
}

// DeletePosition — удалить должность.
func (a *App) DeletePosition(positionId int64) (bool, error) {
	return handlers.DeletePosition(positionId)
}

// === Сотрудники ===

// GetStaff — список сотрудников с подтянутым названием должности.
func (a *App) GetStaff() ([]models.StaffWithPosition, error) {
	return handlers.GetStaff()
}

// CreateStaff — добавить сотрудника.
func (a *App) CreateStaff(fullName string, positionId int64) (models.StaffWithPosition, error) {
	return handlers.CreateStaff(fullName, positionId)
}

// UpdateStaff — обновить ФИО/должность сотрудника.
func (a *App) UpdateStaff(staffId int64, newFullName string, newPositionId int64) (models.StaffWithPosition, error) {
	return handlers.UpdateStaff(staffId, newFullName, newPositionId)
}

// DeleteStaff — удалить сотрудника.
func (a *App) DeleteStaff(staffId int64) (bool, error) {
	return handlers.DeleteStaff(staffId)
}

// === Сервисная команда (использовалась в Tauri-версии для проверки IPC) ===

// Greet возвращает приветствие. Оставлено для отладки/обратной совместимости.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Привет, %s! Бекенд OpticERP работает.", name)
}
