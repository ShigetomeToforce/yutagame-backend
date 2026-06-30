package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"yutagame-backend/application/usecase"
	"yutagame-backend/infrastructure/database"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. データベース接続情報 (コンテナ環境変数から取得)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// 2. GORMの初期化接続
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 3. レイヤーの組み立て (Dependency Injection)
	// --- Repository 層 ---
	machineRepo := database.NewMachineRepository(db)
	gameRepo := database.NewGameRepository(db)
	keywordRepo := database.NewKeywordRepository(db)
	genreRepo := database.NewGenreRepository(db)
	manufacturerRepo := database.NewManufacturerRepository(db)
	_ = database.NewAdminRepository(db) // 認証フェーズで使用するため初期化の準備だけ
	_ = database.NewUserRepository(db)  // 同上

	// --- UseCase 層 ---
	machineUseCase := usecase.NewMachineUseCase(machineRepo)
	gameUseCase := usecase.NewGameUseCase(gameRepo)
	keywordUseCase := usecase.NewKeywordUseCase(keywordRepo)
	genreUseCase := usecase.NewGenreUseCase(genreRepo)
	manufacturerUseCase := usecase.NewManufacturerUseCase(manufacturerRepo)

	// --- Handler 層 ---
	machineHandler := handler.NewMachineHandler(machineUseCase)
	gameHandler := handler.NewGameHandler(gameUseCase)
	keywordHandler := handler.NewKeywordHandler(keywordUseCase)
	genreHandler := handler.NewGenreHandler(genreUseCase)
	manufacturerHandler := handler.NewManufacturerHandler(manufacturerUseCase)

	// 4. Echo インスタンスの生成と共通設定
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// 💡 画像ファイルの静的配信を有効化 (/images/games/1.jpg などでアクセス可能に)
	e.Static("/images", "storage")

	// 5. ルーティング定義
	api := e.Group("/api")
	{
		// 🎮 ゲーム関連API
		api.GET("/games", gameHandler.Search)
		api.GET("/games/:id", gameHandler.GetByID)
		api.POST("/games", gameHandler.Create)

		// 💻 機種関連API
		api.GET("/machines", machineHandler.GetAll)
		api.GET("/machines/:id", machineHandler.GetByID)
		api.POST("/machines", machineHandler.Create)

		// 🏷️ キーワード関連API
		api.GET("/keywords", keywordHandler.GetAll)

		// 🧬 ジャンル関連API
		api.GET("/genres", genreHandler.GetAll)

		// 🏭 メーカー関連API
		api.GET("/manufacturers", manufacturerHandler.GetAll)
	}

	// 6. サーバー起動
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
