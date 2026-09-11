# YUTAGAME Backend

所持ゲーム管理システム（YUTAGAME）のバックエンドAPIサーバーです。

## 技術スタック

- Go: コンパイル時に型を検査するサーバー向け言語
- Echo: HTTPルーティングとmiddlewareを提供するWebフレームワーク
- GORM: Goの構造体とMySQLのテーブルを対応させるORM
- MySQL 8: ゲーム、管理者、アクセス集計などの永続化
- JWT: 管理APIの認証
- Swagger: API仕様の閲覧と動作確認

## ディレクトリ

- `domain/model`: DBテーブルに対応するデータ構造
- `infrastructure/database`: GORMによる検索・保存
- `application/usecase`: 検索条件や重複防止などの業務ルール
- `interface/handler`: HTTP入力の変換とレスポンス
- `interface/middleware`: 認証・ログなど全リクエスト共通処理
- `migrations`: 本番DBにも順番に適用するSQL
- `docs`: Swaggerの自動生成物

処理は基本的に `Handler -> UseCase -> Repository -> MySQL` の順に進みます。
詳細は [DEVELOPMENT_CONCEPT.md](DEVELOPMENT_CONCEPT.md) を参照してください。

## 起動

Docker Desktopを起動してから実行します。

```bash
docker compose up -d --build
docker compose logs -f
docker compose down
```

Goプロセスはソース変更を自動再起動しません。変更を反映するには次を実行します。

```bash
docker compose restart backend
```

## DB migration

`migrations` のSQLは番号順に一度だけ適用されます。

```bash
go run ./cmd/migrate up
```

既存migrationは変更せず、新しい変更は次の連番ファイルへ追加します。
このコマンドは新規環境とgoose管理済みの本番環境向けです。goose導入前から使っているローカルDBは適用履歴がなく、`000001`
と既存テーブルが衝突します。ローカルの `docker-compose.yml` は
`RUN_AUTO_MIGRATE=true`
のため、通常はバックエンド再起動時のAutoMigrateでスキーマを同期します。

## 開発コマンド

```bash
gofmt -w ./...
go test ./...
go build ./...
```

Swaggerを注釈から再生成します。

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go --output docs
```

起動後のSwagger UIは `http://localhost:8080/swagger/index.html` です。

## APIと認証

- `/api/app/*`: エンドユーザー向け公開API
- `/api/admin/login`: 管理者ログイン
- `/api/admin/*`: JWTが必要な管理API

管理APIは `Authorization: Bearer <token>`
を検証します。サーバー側にログインセッションを保持しないStateless方式です。

## PV・UU

公開HTMLページの正常なGETを `/api/app/page-views` へ記録します。

- PV: 同一訪問者・同一ページはJSTの1日につき1回
- 日次UU: その日に閲覧した訪問者をページ横断で1回
- 月次UU: その月に閲覧した訪問者を月全体で1回
- 識別: `visitor_id` Cookieを優先し、保存時はSHA-256ハッシュ化
- 重複防止: `visitor_hash + page_path + viewed_on` のDB一意制約

ゲーム別閲覧ランキングは `game_view_logs`、サイト全体のPV/UUは `page_view_logs`
と役割を分けています。

## ローカルメール確認

お問い合わせ通知メールとおすすめゲーム投稿通知メールはSMTPで送信します。Docker環境ではMailpitを同時に起動し、外部に送信せずローカルで確認できます。

```bash
docker compose up -d --build
```

- SMTP: `mailpit:1025`
- 確認画面: `http://localhost:8025`
- 通知先: `CONTACT_NOTIFY_TO`

本番環境では `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` /
`SMTP_FROM` / `CONTACT_NOTIFY_TO` / `ADMIN_SITE_URL`
を環境変数で設定してください。
