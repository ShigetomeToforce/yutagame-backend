package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/infrastructure/database"
	handlerAdmin "yutagame-backend/interface/handler/admin"  // 💡 エイリアスを付けてインポート
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
	userRepo := database.NewUserRepository(db)

	// --- UseCase 層 ---
	machineUseCase := usecaseAdmin.NewMachineUseCase(machineRepo)
	gameUseCase := usecaseAdmin.NewGameUseCase(gameRepo)
	keywordUseCase := usecaseAdmin.NewKeywordUseCase(keywordRepo)
	genreUseCase := usecaseAdmin.NewGenreUseCase(genreRepo)
	manufacturerUseCase := usecaseAdmin.NewManufacturerUseCase(manufacturerRepo)
	adminUseCase := usecaseAdmin.NewAdminUseCase(adminRepo)
	userUseCase := usecaseAdmin.NewUserUseCase(userRepo)

	// --- Handler 層 ---
	machineHandler := handlerAdmin.NewMachineHandler(machineUseCase)
	gameHandler := handlerAdmin.NewGameHandler(gameUseCase)
	keywordHandler := handlerAdmin.NewKeywordHandler(keywordUseCase)
	genreHandler := handlerAdmin.NewGenreHandler(genreUseCase)
	manufacturerHandler := handlerAdmin.NewManufacturerHandler(manufacturerUseCase)
	adminHandler := handlerAdmin.NewAdminHandler(adminUseCase)
	userHandler := handlerAdmin.NewUserHandler(userUseCase)

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
		api.POST("/admin/login", adminHandler.Login)

		// 🔒 【認証必須エリア：管理画面専用】
		// 💡 GETも含め、現在実装されている生のリソースAPIはすべてガードの中に幽閉します
		adminProtected := api.Group("/admin")
		adminProtected.Use(customMiddleware.AdminGuard()) // 自作の認証ミドルウェア
		{
			// 👥 Adminユーザー管理
			adminProtected.GET("/admins", adminHandler.GetAll)
			adminProtected.GET("/admins/:id", adminHandler.GetByID)
			adminProtected.POST("/admins", adminHandler.Create)
			adminProtected.PUT("/admins/:id", adminHandler.Update)
			adminProtected.DELETE("/admins/:id", adminHandler.Delete)

			// 🎮 ゲーム管理
			adminProtected.GET("/games", gameHandler.GetAll)
			adminProtected.GET("/games/:id", gameHandler.GetByID)
			adminProtected.POST("/games", gameHandler.Create)
			adminProtected.PUT("/games/:id", gameHandler.Update)
			adminProtected.DELETE("/games/:id", gameHandler.Delete)

			// 💻 機種管理
			adminProtected.GET("/machines", machineHandler.GetAll)
			adminProtected.GET("/machines/:id", machineHandler.GetByID)
			adminProtected.POST("/machines", machineHandler.Create)
			adminProtected.PUT("/machines/:id", machineHandler.Update)
			adminProtected.DELETE("/machines/:id", machineHandler.Delete)

			// 🏷️ キーワード管理
			adminProtected.GET("/keywords", keywordHandler.GetAll)
			adminProtected.GET("/keywords/:id", keywordHandler.GetByID)
			adminProtected.POST("/keywords", keywordHandler.Create)
			adminProtected.PUT("/keywords/:id", keywordHandler.Update)
			adminProtected.DELETE("/keywords/:id", keywordHandler.Delete)

			// 🧬 ジャンル管理
			adminProtected.GET("/genres", genreHandler.GetAll)
			adminProtected.GET("/genres/:id", genreHandler.GetByID)
			adminProtected.POST("/genres", genreHandler.Create)
			adminProtected.PUT("/genres/:id", genreHandler.Update)
			adminProtected.DELETE("/genres/:id", genreHandler.Delete)

			// 🏭 メーカー管理
			adminProtected.GET("/manufacturers", manufacturerHandler.GetAll)
			adminProtected.GET("/manufacturers/:id", manufacturerHandler.GetByID)
			adminProtected.POST("/manufacturers", manufacturerHandler.Create)
			adminProtected.PUT("/manufacturers/:id", manufacturerHandler.Update)
			adminProtected.DELETE("/manufacturers/:id", manufacturerHandler.Delete)

			// 🔓 ユーザー管理
			adminProtected.GET("/users", userHandler.GetAll)
			adminProtected.GET("/users/:id", userHandler.GetByID)
			adminProtected.POST("/users", userHandler.Create)
			adminProtected.PUT("/users/:id", userHandler.Update)
			adminProtected.DELETE("/users/:id", userHandler.Delete)

		}
	}

	// 6. サーバー起動
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
