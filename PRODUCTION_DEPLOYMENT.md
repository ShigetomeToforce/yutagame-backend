# Production Deployment Notes

AWSアカウント作成からWeb公開までの具体的な初心者向け手順は
[AWS_PRODUCTION_BEGINNER_GUIDE.md](AWS_PRODUCTION_BEGINNER_GUIDE.md)
を参照する。

## 方針

本番環境ではDockerを使わず、EC2
1台にアプリケーション、MySQL、リバースプロキシを配置する低コスト構成から始める。

ローカル環境は従来どおり `init/`
のseedデータを使う。本番環境ではseedデータを流さず、goose
migrationで空DBのテーブルを作成し、デフォルト管理者だけを冪等に投入する。

## 本番DB初期化

本番では `cmd/migrate` から以下を実行する。

- gooseによるSQL migration適用
- `admins.id = 1` のデフォルト管理者作成（既存レコードは上書きしない）

デフォルト管理者は以下の環境変数で上書きできる。

- `DEFAULT_ADMIN_EMAIL`
- `DEFAULT_ADMIN_PASSWORD_HASH`
- `DEFAULT_ADMIN_NAME`
- `DEFAULT_ADMIN_ROLE`

未設定時はローカルseedと同じ管理者が作成される。

CI/CDや初回セットアップでは、backendとは別にmigration用バイナリをビルドして実行する。

```bash
go build -o yutagame-migrate ./cmd/migrate
./yutagame-migrate up
```

## 最小AWS構成

- EC2: `t4g.micro` または無料枠/低額インスタンス
- OS: Ubuntu LTS
- DB: 同一EC2内のMySQL 8
- Web: Caddy または Nginx
- SSL: Let's Encrypt
- DNS: Route 53
- メール: Amazon SES SMTP
- バックアップ: `mysqldump` をS3へ日次保存
- 監視: CloudWatch Agent + ディスク使用率監視
- デプロイ: GitHub ActionsからSSH、またはSelf-hosted runner

単一EC2構成は安いが、EC2障害時にアプリとDBが同時に止まる。最初はこの構成で始め、アクセスや売上が増えたらRDS、ALB、複数台構成へ移行する。

## SSL

無料で運用するならCaddyが最も簡単。CaddyはLet's
Encrypt証明書の取得と更新を自動化できる。

例:

```caddyfile
example.com {
  reverse_proxy 127.0.0.1:8000
}

api.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

複数サイトを同じEC2で運用する場合も、サブドメインごとにreverse_proxyを追加する。

## バックアップ

最低限、DBは日次でS3へ退避する。

```bash
#!/usr/bin/env bash
set -euo pipefail

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/var/backups/yutagame
mkdir -p "$BACKUP_DIR"

mysqldump -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" | gzip > "$BACKUP_DIR/yutagame_$DATE.sql.gz"
aws s3 cp "$BACKUP_DIR/yutagame_$DATE.sql.gz" "s3://your-backup-bucket/yutagame/"
find "$BACKUP_DIR" -type f -name 'yutagame_*.sql.gz' -mtime +7 -delete
```

S3側はライフサイクルで30日後Glacier、90日後削除などにすると安い。バックアップは取得だけでなく、定期的に復元確認する。

## メール

本番はAmazon SES
SMTPを推奨する。低コストで、問い合わせ通知/おすすめゲーム投稿通知の用途に合う。

必要な環境変数:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM`
- `CONTACT_NOTIFY_TO`
- `ADMIN_SITE_URL`

SESは最初サンドボックス状態なので、本番送信前に解除申請が必要。

## 機密情報

DBパスワード、JWT secret、SMTPパスワードはGitHubに置かない。

低コスト優先ならEC2上の環境ファイルで管理する。

```text
/etc/yutagame/backend.env
/etc/yutagame/frontend.env
```

権限は `600`、所有者はアプリ実行ユーザーに限定する。より堅くするならAWS Systems
Manager Parameter Storeを使う。

## systemd

バックエンドはsystemdで常駐させる。

```ini
[Unit]
Description=Yutagame Backend
After=network.target mysql.service

[Service]
WorkingDirectory=/opt/yutagame/yutagame-backend
EnvironmentFile=/etc/yutagame/backend.env
ExecStart=/opt/yutagame/yutagame-backend/yutagame-backend
Restart=always
RestartSec=5
User=yutagame

[Install]
WantedBy=multi-user.target
```

フロントエンドも同様に `deno run -A main.ts`
またはビルド成果物をsystemdで起動する。

## CI/CD

低コストで始めるならGitHub ActionsからSSHでデプロイする。

流れ:

1. GitHub ActionsでGoビルド、Denoビルド/チェック
2. EC2へ成果物をrsync/scp
3. `systemctl restart yutagame-backend yutagame-frontend`

より安全にするならEC2上にSelf-hosted
runnerを置く。ただしrunner権限管理に注意する。

## 監視

最低限見るもの:

- CPU使用率
- メモリ使用率
- ディスク使用率
- MySQLプロセス
- backend/frontendのsystemd status
- 5xxログ件数

CloudWatch
AgentでCPU/メモリ/ディスクを送り、ディスク80%超過でアラームを設定する。

## 複数サイト運用

同じEC2内で複数サイトを運用する場合:

- サブドメイン単位でCaddy/Nginxのreverse_proxyを分ける
- MySQLはDBスキーマを分ける
- Linuxユーザーとsystemd serviceをサイトごとに分ける
- `/etc/{site-name}/backend.env` のように環境変数も分ける
- ログディレクトリも `/var/log/{site-name}/...` に分ける

将来的に別インスタンスへ分ける可能性があるなら、OpenTofuで以下を管理する。

- VPC/Security Group
- EC2
- Elastic IP
- Route 53 record
- S3 backup bucket
- IAM role

最初から全部をIaC化しすぎると重くなるため、まずはEC2/Security Group/Route
53/S3だけをOpenTofu化するのが現実的。

## マイグレーション管理

本番はgooseでSQL migrationを管理する。適用済みversionはDB内の `goose_db_version`
に保存される。

基本操作:

```bash
./yutagame-migrate status
./yutagame-migrate up
./yutagame-migrate down
```

本番では `down` はバックアップ取得後に慎重に使う。
