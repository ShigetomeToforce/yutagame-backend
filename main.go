package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
	handlerAdmin "yutagame-backend/interface/handler/admin" // 💡 エイリアスを付けてインポート
	handlerApp "yutagame-backend/interface/handler/app"
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
// @description     所持ゲーム管理システムのバックエンドAPI仕様書
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

	if err := db.AutoMigrate(&model.Announcement{}, &model.ContactInquiry{}); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
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
	announcementRepo := database.NewAnnouncementRepository(db)
	contactRepo := database.NewContactInquiryRepository(db)

	// --- UseCase 層 ---
	machineUseCase := usecaseAdmin.NewMachineUseCase(machineRepo)
	gameUseCase := usecaseAdmin.NewGameUseCase(gameRepo)
	keywordUseCase := usecaseAdmin.NewKeywordUseCase(keywordRepo)
	genreUseCase := usecaseAdmin.NewGenreUseCase(genreRepo)
	manufacturerUseCase := usecaseAdmin.NewManufacturerUseCase(manufacturerRepo)
	adminUseCase := usecaseAdmin.NewAdminUseCase(adminRepo)
	userUseCase := usecaseAdmin.NewUserUseCase(userRepo)
	announcementUseCase := usecaseAdmin.NewAnnouncementUseCase(announcementRepo)
	contactUseCase := usecaseAdmin.NewContactInquiryUseCase(contactRepo)
	publicGameUseCase := usecaseApp.NewGamePublicUseCase(
		gameRepo,
		machineRepo,
		genreRepo,
		manufacturerRepo,
		keywordRepo,
	)
	publicAnnouncementUseCase := usecaseApp.NewAnnouncementPublicUseCase(announcementRepo)
	publicContactUseCase := usecaseApp.NewContactPublicUseCase(contactRepo)
	siteUseCase := usecaseApp.NewSiteUseCase(gameRepo, announcementRepo)

	// --- Handler 層 ---
	machineHandler := handlerAdmin.NewMachineHandler(machineUseCase)
	gameHandler := handlerAdmin.NewGameHandler(gameUseCase)
	keywordHandler := handlerAdmin.NewKeywordHandler(keywordUseCase)
	genreHandler := handlerAdmin.NewGenreHandler(genreUseCase)
	manufacturerHandler := handlerAdmin.NewManufacturerHandler(manufacturerUseCase)
	adminHandler := handlerAdmin.NewAdminHandler(adminUseCase)
	userHandler := handlerAdmin.NewUserHandler(userUseCase)
	announcementHandler := handlerAdmin.NewAnnouncementHandler(announcementUseCase)
	contactInquiryHandler := handlerAdmin.NewContactInquiryHandler(contactUseCase)
	publicGameHandler := handlerApp.NewGameHandler(publicGameUseCase)
	publicAnnouncementHandler := handlerApp.NewAnnouncementHandler(publicAnnouncementUseCase)
	publicContactHandler := handlerApp.NewContactHandler(publicContactUseCase)
	siteHandler := handlerApp.NewSiteHandler(siteUseCase)

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
		// 🌐 【公開エリア】一般公開向けの参照系API
		public := api.Group("/app")
		{
			public.GET("/top", publicGameHandler.GetTop)
			public.GET("/announcements", publicAnnouncementHandler.GetAll)
			public.GET("/announcements/:id", publicAnnouncementHandler.GetByID)
			public.POST("/contacts", publicContactHandler.Create)
			public.GET("/sitemap", siteHandler.GetSitemap)
			public.GET("/catalog/machines", publicGameHandler.GetMachines)
			public.GET("/catalog/genres", publicGameHandler.GetGenres)
			public.GET("/catalog/manufacturers", publicGameHandler.GetManufacturers)
			public.GET("/keywords", publicGameHandler.GetKeywords)
			public.GET("/games", publicGameHandler.Search)
			public.GET("/games/:code", publicGameHandler.GetByCode)
		}

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
			adminProtected.GET("/games/code/:code", gameHandler.GetByCode)
			adminProtected.GET("/games/:id", gameHandler.GetByID)
			adminProtected.POST("/games", gameHandler.Create)
			adminProtected.PUT("/games/:id", gameHandler.Update)
			adminProtected.DELETE("/games/:id", gameHandler.Delete)
			adminProtected.POST("/games/export", gameHandler.ExportCSV)
			adminProtected.POST("/games/import/preview", gameHandler.PreviewImportCSV)
			adminProtected.POST("/games/import/apply", gameHandler.ApplyImportCSV)
			adminProtected.POST("/games/:id/image", gameHandler.UploadImage)
			adminProtected.DELETE("/games/:id/image", gameHandler.DeleteImage)
			adminProtected.GET("/games/:id/affiliates", gameHandler.ListAffiliates)
			adminProtected.POST("/games/:id/affiliates", gameHandler.CreateAffiliate)
			adminProtected.PUT("/games/:id/affiliates/:affiliateId", gameHandler.UpdateAffiliate)
			adminProtected.DELETE("/games/:id/affiliates/:affiliateId", gameHandler.DeleteAffiliate)

			// 💻 機種管理
			adminProtected.GET("/machines", machineHandler.GetAll)
			adminProtected.GET("/machines/code/:code", machineHandler.GetByCode)
			adminProtected.GET("/machines/:id", machineHandler.GetByID)
			adminProtected.POST("/machines", machineHandler.Create)
			adminProtected.PUT("/machines/:id", machineHandler.Update)
			adminProtected.DELETE("/machines/:id", machineHandler.Delete)
			adminProtected.POST("/machines/export", machineHandler.ExportCSV)
			adminProtected.POST("/machines/import/preview", machineHandler.PreviewImportCSV)
			adminProtected.POST("/machines/import/apply", machineHandler.ApplyImportCSV)
			adminProtected.POST("/machines/:id/image", machineHandler.UploadImage)
			adminProtected.DELETE("/machines/:id/image", machineHandler.DeleteImage)

			// 🏷️ キーワード管理
			adminProtected.GET("/keywords", keywordHandler.GetAll)
			adminProtected.GET("/keywords/code/:code", keywordHandler.GetByCode)
			adminProtected.GET("/keywords/:id", keywordHandler.GetByID)
			adminProtected.POST("/keywords", keywordHandler.Create)
			adminProtected.PUT("/keywords/:id", keywordHandler.Update)
			adminProtected.DELETE("/keywords/:id", keywordHandler.Delete)
			adminProtected.POST("/keywords/export", keywordHandler.ExportCSV)
			adminProtected.POST("/keywords/import/preview", keywordHandler.PreviewImportCSV)
			adminProtected.POST("/keywords/import/apply", keywordHandler.ApplyImportCSV)

			adminProtected.GET("/announcements", announcementHandler.GetAll)
			adminProtected.GET("/announcements/:id", announcementHandler.GetByID)
			adminProtected.POST("/announcements", announcementHandler.Create)
			adminProtected.PUT("/announcements/:id", announcementHandler.Update)
			adminProtected.DELETE("/announcements/:id", announcementHandler.Delete)

			adminProtected.GET("/contacts", contactInquiryHandler.GetAll)
			adminProtected.GET("/contacts/:id", contactInquiryHandler.GetByID)
			adminProtected.PUT("/contacts/:id", contactInquiryHandler.Update)
			adminProtected.DELETE("/contacts/:id", contactInquiryHandler.Delete)

			// 🧬 ジャンル管理
			adminProtected.GET("/genres", genreHandler.GetAll)
			adminProtected.GET("/genres/code/:code", genreHandler.GetByCode)
			adminProtected.GET("/genres/:id", genreHandler.GetByID)
			adminProtected.POST("/genres", genreHandler.Create)
			adminProtected.PUT("/genres/:id", genreHandler.Update)
			adminProtected.DELETE("/genres/:id", genreHandler.Delete)
			adminProtected.POST("/genres/export", genreHandler.ExportCSV)
			adminProtected.POST("/genres/import/preview", genreHandler.PreviewImportCSV)
			adminProtected.POST("/genres/import/apply", genreHandler.ApplyImportCSV)
			adminProtected.POST("/genres/:id/image", genreHandler.UploadImage)
			adminProtected.DELETE("/genres/:id/image", genreHandler.DeleteImage)

			// 🏭 メーカー管理
			adminProtected.GET("/manufacturers", manufacturerHandler.GetAll)
			adminProtected.GET("/manufacturers/code/:code", manufacturerHandler.GetByCode)
			adminProtected.GET("/manufacturers/:id", manufacturerHandler.GetByID)
			adminProtected.POST("/manufacturers", manufacturerHandler.Create)
			adminProtected.PUT("/manufacturers/:id", manufacturerHandler.Update)
			adminProtected.DELETE("/manufacturers/:id", manufacturerHandler.Delete)
			adminProtected.POST("/manufacturers/export", manufacturerHandler.ExportCSV)
			adminProtected.POST("/manufacturers/import/preview", manufacturerHandler.PreviewImportCSV)
			adminProtected.POST("/manufacturers/import/apply", manufacturerHandler.ApplyImportCSV)
			adminProtected.POST("/manufacturers/:id/image", manufacturerHandler.UploadImage)
			adminProtected.DELETE("/manufacturers/:id/image", manufacturerHandler.DeleteImage)

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
