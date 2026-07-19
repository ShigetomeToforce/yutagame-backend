package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	adminUsecase "yutagame-backend/application/usecase/admin"
	"yutagame-backend/infrastructure/database"
	adminHandler "yutagame-backend/interface/handler/admin"  // 💡 エイリアスを付けてインポート
	customMiddleware "yutagame-backend/interface/middleware" // 💡 追加

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "yutagame-backend/docs" // 💡 swag init で自動生成されるドキュメントを読み込む

	echoSwagger "github.com/swaggo/echo-swagger" //
)

// @title           Yutagame Backend API
// @version         1.0
// @description     ゲーム在庫管理システムのバックエンドAPI仕様書
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                "Bearer {token}" の形式でJWTトークンを入力してください。
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
	adminRepo := database.NewAdminRepository(db)
	_ = database.NewUserRepository(db) // 将来の一般ユーザー用（準備だけ）

	// --- UseCase 層 ---
	machineUseCase := adminUsecase.NewMachineUseCase(machineRepo)
	gameUseCase := adminUsecase.NewGameUseCase(gameRepo)
	keywordUseCase := adminUsecase.NewKeywordUseCase(keywordRepo)
	genreUseCase := adminUsecase.NewGenreUseCase(genreRepo)
	manufacturerUseCase := adminUsecase.NewManufacturerUseCase(manufacturerRepo)
	adminAuthUseCase := adminUsecase.NewAdminAuthUseCase(adminRepo)

	// --- Handler 層 ---
	machineHandler := adminHandler.NewMachineHandler(machineUseCase)
	gameHandler := adminHandler.NewGameHandler(gameUseCase)
	keywordHandler := adminHandler.NewKeywordHandler(keywordUseCase)
	genreHandler := adminHandler.NewGenreHandler(genreUseCase)
	manufacturerHandler := adminHandler.NewManufacturerHandler(manufacturerUseCase)
	adminAuthHandler := adminHandler.NewAdminAuthHandler(adminAuthUseCase)

	// 4. Echo インスタンスの生成と共通設定
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}, // 💡 Authorizationヘッダーを許可
	}))

	// 画像ファイルの静的配信を有効化
	e.Static("/images", "storage")

	// 💡 Swagger UI のエンドポイントを追加 (認証なしで誰でも見られる場所)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// 5. ルーティング定義
	api := e.Group("/api")
	{
		// 🔓 【完全公開エリア】ログインAPIのみ外に出す
		api.POST("/admin/login", adminAuthHandler.Login)

		// 🔒 【認証必須エリア：管理画面専用】
		// 💡 GETも含め、現在実装されている生のリソースAPIはすべてガードの中に幽閉します
		adminProtected := api.Group("/admin")
		adminProtected.Use(customMiddleware.AdminGuard()) // 自作の認証ミドルウェア
		{
			// 👥 管理者アカウント自体の管理 (CRUD)
			adminProtected.GET("/admins", adminAuthHandler.GetAll)
			adminProtected.GET("/admins/:id", adminAuthHandler.GetByID)
			adminProtected.POST("/admins", adminAuthHandler.Create)
			adminProtected.PUT("/admins/:id", adminAuthHandler.Update)
			adminProtected.DELETE("/admins/:id", adminAuthHandler.Delete)

			// 🎮 ゲーム管理
			adminProtected.GET("/games", gameHandler.Search)
			adminProtected.GET("/games/:id", gameHandler.GetByID)
			adminProtected.POST("/games", gameHandler.Create)

			// 💻 機種管理
			adminProtected.GET("/machines", machineHandler.GetAll)
			adminProtected.GET("/machines/:id", machineHandler.GetByID)
			adminProtected.POST("/machines", machineHandler.Create)

			// 🏷️ キーワード管理
			adminProtected.GET("/keywords", keywordHandler.GetAll)

			// 🧬 ジャンル管理
			adminProtected.GET("/genres", genreHandler.GetAll)
			adminProtected.GET("/genres/:id", genreHandler.GetByID)
			adminProtected.POST("/genres", genreHandler.Create)
			adminProtected.PUT("/genres/:id", genreHandler.Update)
			adminProtected.DELETE("/genres/:id", genreHandler.Delete)

			// 🏭 メーカー管理
			adminProtected.GET("/manufacturers", manufacturerHandler.GetAll)
		}
	}

	// 6. サーバー起動
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
