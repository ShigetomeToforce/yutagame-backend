# YUTAGAME Backend (Go + Echo)

ゲーム在庫管理システム（YUTAGAME）のバックエンドAPIサーバーです。

## 🛠️ 技術スタック

- **言語**: Go
- **Webフレームワーク**: Echo
- **データベース**: MySQL (Docker管理)
- **認証方式**: JWT (JSON Web Token)
- **実行環境**: Docker / Docker Compose

## 🔑 認証設計のポイント

- **状態を持たない（Stateless）認証**:
  サーバー側でセッションを保持せず、フロントエンドからリクエストごとに送られてくるJWT（トークン）を検証する設計です。
- **トークン検証ルール**:
  リクエストの `Authorization` ヘッダーに含まれる `Bearer [トークン]` を解析し、有効期限（72時間）および署名の正当性をチェックします。

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
