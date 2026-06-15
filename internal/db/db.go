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
		// modernc.org/sqlite на Windows лучше работает с прямым путём без file://
		dsn := dbPath + "?_pragma=foreign_keys(1)"
		conn, err := sql.Open("sqlite", dsn)
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

// DBPath возвращает путь к erp.db.
//
// По требованию проекта БД должна лежать в корне репозитория
// (рядом с go.mod), чтобы её удобно было открывать в IDE/расширениях
// и держать в git-ignore. То есть:
//
//	f:\OpticERP\erp.db
//
// Эта логика совпадает с поведением оригинальной Tauri-версии, где
// Cargo-бинарь собирался в корень проекта и erp.db создавалась рядом.
//
// Приоритеты:
//  1. Переменная окружения OPTICERP_DB_PATH — для тестов и отладки.
//  2. <корень проекта>/erp.db — определяется по наличию go.mod в
//     текущей рабочей директории или её родителях. Это нужно, потому
//     что в `wails dev` рабочая директория обычно равна `frontend/`,
//     а exe — `build/bin/OpticERP.exe`. Идём вверх по дереву, пока не
//     найдём go.mod.
//  3. Фоллбэк — текущая рабочая директория.
func DBPath() string {
	if envPath := os.Getenv("OPTICERP_DB_PATH"); envPath != "" {
		return envPath
	}
	if projectRoot := findProjectRoot(); projectRoot != "" {
		return filepath.Join(projectRoot, "erp.db")
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, "erp.db")
}

// findProjectRoot поднимается вверх по дереву директорий от текущей
// рабочей директории (или от директории exe, если она задана) и
// возвращает путь к директории, содержащей go.mod. Если ничего не
// найдено — возвращает пустую строку.
func findProjectRoot() string {
	candidates := []string{}
	if wd, err := os.Getwd(); err == nil && wd != "" {
		candidates = append(candidates, wd)
	}
	if exe, err := os.Executable(); err == nil && exe != "" {
		candidates = append(candidates, filepath.Dir(exe))
	}
	for _, start := range candidates {
		dir := start
		for i := 0; i < 8; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}
