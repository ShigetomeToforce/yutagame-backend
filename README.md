# YUTAGAME Backend (Go + Echo)

所持ゲーム管理システム（YUTAGAME）のバックエンドAPIサーバーです。

## 🛠️ 技術スタック

- **言語**: Go
- **Webフレームワーク**: Echo
- **データベース**: MySQL (Docker管理)
- **認証方式**: JWT (JSON Web Token)
- **実行環境**: Docker / Docker Compose

## 🔑 認証設計のポイント

- **状態を持たない（Stateless）認証**:
  サーバー側でセッションを保持せず、フロントエンドからリクエストごとに送られてくるJWT（トークン）を検証する設計です。
- **トークン検証ルール**: リクエストの `Authorization` ヘッダーに含まれる
  `Bearer [トークン]`
  を解析し、有効期限（72時間）および署名の正当性をチェックします。

## 🚀 起動方法（Docker環境）

ローカル開発環境およびデータベースは Docker Compose を使用して起動します。

```bash
# コンテナのビルドと起動（バックグラウンド実行）
docker compose up -d

# ログの確認
docker compose logs -f

# コンテナの停止
docker compose down
```

## 📧 ローカルメール確認

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
