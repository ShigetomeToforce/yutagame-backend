package main

import (
	"fmt"
	"net/http"

	// MySQLの通訳ライブラリ（直接コードでは使わないが、裏側で登録するために _ をつけてインポートします）
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func main() {
	// 1. MySQLデータベースへの接続設定（ユーザー名:パスワード@接続先/データベース名）
	// docker-compose.yml で設定した値と合わせています
	dataSourceName := "yuta_user:yuta_password@tcp(localhost:3306)/yutagame?parseTime=true"

	// 2. データベースを開く
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(fmt.Sprintf("データベースを開けませんでした: %v", err))
	}
	defer db.Close() // プログラム終了時に安全にクローズする設定

	// 3. 実際に繋がるかテスト（Pingを打つ）
	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("MySQLへの接続テストに失敗しました: %v", err))
	}

	// 接続成功メッセージをターミナルに出す
	fmt.Println("🎉 MySQLデータベースへの接続に成功しました！")

	// --- ここからは前回と同じEchoのWebサーバー設定 ---
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, YutaGame API (Database Connected!)")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
