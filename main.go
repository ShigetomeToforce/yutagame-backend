package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"yutagame-backend/application/usecase"
	"yutagame-backend/infrastructure/database"
	"yutagame-backend/interface/handler"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 💡 追加: .env ファイルが存在する場合のみ環境変数をロードする
	// (管理画面などの本番環境で、コンテナ側に直接環境変数を注入する場合にもバグらない書き方です)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	// 1. データベース接続情報 (環境変数から取得)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// 2. GORMの初期化接続
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 💡 ターミナルにGORMが自動生成したクリーンなSQLが流れます
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 3. レイヤーの組み立て (Dependency Injection)
	// --- Repository 層 ---
	machineRepo := database.NewMachineRepository(db)
	gameRepo := database.NewGameRepository(db)
	keywordRepo := database.NewKeywordRepository(db)

	// --- UseCase (Service) 層 ---
	machineUseCase := usecase.NewMachineUseCase(machineRepo)
	gameUseCase := usecase.NewGameUseCase(gameRepo)
	keywordUseCase := usecase.NewKeywordUseCase(keywordRepo)

	// --- Handler (Controller) 層 ---
	machineHandler := handler.NewMachineHandler(machineUseCase)
	gameHandler := handler.NewGameHandler(gameUseCase)
	keywordHandler := handler.NewKeywordHandler(keywordUseCase)

	// 4. Echo インスタンスの生成と共通設定
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// 💡 コンテナ内の /app/storage フォルダを /images というURLパスで外部公開する
	e.Static("/images", "storage")

	// 5. ルーティング定義 (URLとWeb窓口の紐付け)
	api := e.Group("/api")
	{
		// 🎮 ゲーム関連API
		api.GET("/games", gameHandler.Search)
		api.GET("/games/:id", gameHandler.GetByID)
		api.POST("/games", gameHandler.Create) // 将来の管理用

		// 💻 機種関連API
		api.GET("/machines", machineHandler.GetAll)
		api.GET("/machines/:id", machineHandler.GetByID)
		api.POST("/machines", machineHandler.Create) // 将来の管理用

		// 🏷️ キーワード関連API
		api.GET("/keywords", keywordHandler.GetAll)
	}

	// 6. サーバー起動
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
