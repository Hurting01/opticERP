package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"OpticERP/internal/db"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Подключаемся к БД и применяем миграции ДО запуска UI, чтобы любые
	// ошибки схемы всплыли в логах. Используем log.Print вместо Fatalf,
	// чтобы при сбое БД приложение всё равно запустилось: wails-биндинги
	// генерируются через запуск бинаря в build/bin/, и фатальный exit там
	// ломает процесс сборки. Если БД недоступна — UI стартует без данных,
	// пользователь увидит проблему в логах / на пустых списках.
	conn, err := db.DB()
	if err != nil {
		log.Printf("⚠️  Не удалось подключиться к БД (%s): %v", db.DBPath(), err)
	} else if err := db.Migrate(conn); err != nil {
		log.Printf("⚠️  Ошибка миграции (%s): %v", db.DBPath(), err)
	} else {
		log.Printf("✅ БД инициализирована: %s", db.DBPath())
	}

	app := NewApp()

	err = wails.Run(&options.App{
		Title:  "ERP для салона оптики",
		Width:  1700,
		Height: 950,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 238, G: 242, B: 246, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			runtime.LogInfo(ctx, "OpticERP запущен")
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Printf("Ошибка при запуске приложения: %v", err)
	}
}
