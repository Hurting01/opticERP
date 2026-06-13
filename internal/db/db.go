// Package db отвечает за подключение к локальной SQLite-базе данных
// (erp.db) и применение миграций. Поведение 1:1 соответствует
// оригинальному Tauri-проекту opticTauri (см. src-tauri/src/lib.rs).
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	// Регистрация драйвера modernc.org/sqlite (чистый Go, без CGO).
	_ "modernc.org/sqlite"
)

var (
	dbOnce sync.Once
	dbConn *sql.DB
	dbErr  error
)

// DB возвращает единое соединение с локальной БД.
// Перед использованием вызывается Migrate(), который при необходимости
// пересоздаёт файл erp.db и применяет все миграции.
//
// Путь к БД привязан к директории, где лежит исполняемый файл (а не к
// текущей рабочей директории). Это устраняет зависимость от того, откуда
// пользователь/wails запустил бинарь: dev-режим Wails запускается из
// `frontend/`, а `wails build` кладёт бинарь в `build/bin/`. В обоих
// случаях erp.db окажется рядом с exe, что повторяет поведение
// Tauri-версии, где Cargo-бинарь лежал в project root.
func DB() (*sql.DB, error) {
	dbOnce.Do(func() {
		dbPath := DBPath()
		// Создаём директорию и файл, если их нет (иначе sqlite-соединение
		// будет работать с :memory: или вернёт out-of-memory).
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			dbErr = fmt.Errorf("не удалось создать директорию БД: %w", err)
			return
		}
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			f, err := os.Create(dbPath)
			if err != nil {
				dbErr = fmt.Errorf("не удалось создать файл БД: %w", err)
				return
			}
			_ = f.Close()
		}

		// Параметр _pragma=foreign_keys(on) включает поддержку FK
		// (по умолчанию в SQLite она выключена, что нам не подходит).
		conn, err := sql.Open("sqlite", "file://"+filepath.ToSlash(dbPath)+"?_pragma=foreign_keys(on)")
		if err != nil {
			dbErr = fmt.Errorf("ошибка подключения к БД: %w", err)
			return
		}
		if err := conn.Ping(); err != nil {
			dbErr = fmt.Errorf("ошибка пинга БД: %w", err)
			return
		}
		dbConn = conn
	})
	return dbConn, dbErr
}

// DBPath возвращает абсолютный путь к erp.db рядом с исполняемым файлом.
// Если по какой-то причине нельзя определить путь к exe (например, в
// тестах под go run), откатываемся на текущую рабочую директорию.
func DBPath() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		dir := filepath.Dir(exe)
		if dir != "" {
			return filepath.Join(dir, "erp.db")
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, "erp.db")
}
