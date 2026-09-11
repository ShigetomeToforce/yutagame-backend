# AWS Production Beginner Guide

このドキュメントは、AWSをほぼ初めて触る前提で、`PACKAGE FROESST`
を本番公開するための手順をまとめたものです。

## 0. まず決める構成

最初はコスト優先で、EC2 1台に全部載せる構成にする。

```text
Internet
  |
Route 53 DNS
  |
Elastic IP
  |
EC2 Ubuntu
  |-- Caddy: HTTPS / reverse proxy / Let's Encrypt
  |-- frontend: Fresh / Deno / localhost:8000
  |-- backend: Go / Echo / localhost:8080
  |-- MySQL 8: localhost:3306
  |-- backup script: mysqldump -> S3
  |-- CloudWatch Agent: CPU / memory / disk
```

この構成は安いが、EC2が止まるとWeb/API/DBがすべて止まる。初期運用としては現実的だが、アクセスや売上が増えたらRDS、ALB、複数EC2へ分ける。

## 1. 必要なもの

- AWSアカウント
- GitHubリポジトリ
- 取得したいドメイン名
- クレジットカード
- ローカルPCのターミナル
- SSH鍵

ローカルPCにあると便利なもの:

```bash
aws --version
ssh -V
git --version
```

`aws` コマンドがなければ AWS CLI v2 を入れる。

## 2. AWSアカウント作成

1. AWS公式サイトでアカウント作成
2. ルートユーザーにMFAを設定
3. Billingアラートを設定
4. IAMユーザーまたはIAM Identity Centerを作る
5. 日常作業ではルートユーザーを使わない

最初に必ずやること:

- ルートユーザーMFA
- 請求アラート
- リージョンは `ap-northeast-1`、東京リージョンに統一

## 3. 予算アラート

AWS Consoleで以下を設定する。

1. Billing and Cost Management
2. Budgets
3. Create budget
4. Monthly cost budget
5. 例: `10 USD` か `20 USD`
6. メール通知先を設定

最初は意図しない課金に気づけることが大事。

## 4. Route 53でドメイン取得

1. Route 53を開く
2. Registered domains
3. Register domain
4. ドメインを検索して購入
5. Hosted zone が作られることを確認

例:

```text
example.com
www.example.com
api.example.com
admin.example.com
```

最初は以下のように分けるとわかりやすい。

- `example.com`: 公開サイト
- `api.example.com`: backend API
- `admin.example.com`: 管理画面を同じfrontendへ向ける場合の管理用入口

ただし、このアプリはfrontend側で `/admin` を持っているため、最初は
`example.com/admin` でもよい。

## 5. S3バックアップバケット作成

DBバックアップ保存用のS3バケットを作る。

例:

```text
yutagame-prod-backup-xxxxxxxx
```

設定:

- Block all public access: ON
- Versioning: ON 推奨
- Lifecycle rule:
  - 30日後 Glacier Instant Retrieval か Standard-IA
  - 90日または180日後に削除

バケットは絶対に公開しない。

## 6. EC2用IAM Role作成

EC2からS3へバックアップをアップロードし、CloudWatchへメトリクスを送るためのIAM
Roleを作る。

付ける権限:

- `CloudWatchAgentServerPolicy`
- S3バックアップバケットへの最小権限

S3用のカスタムポリシー例:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": ["s3:PutObject", "s3:ListBucket"],
            "Resource": [
                "arn:aws:s3:::yutagame-prod-backup-xxxxxxxx",
                "arn:aws:s3:::yutagame-prod-backup-xxxxxxxx/*"
            ]
        }
    ]
}
```

## 7. EC2作成

EC2 > Launch instance。

推奨初期設定:

- Name: `yutagame-prod-01`
- AMI: Ubuntu Server 24.04 LTS または 22.04 LTS
- Instance type: `t4g.micro` または `t3.micro`
- Key pair: 新規作成または既存鍵
- Storage: 20GBから30GB gp3
- IAM instance profile: 先ほど作ったRole

Security Group:

```text
Inbound
  22/tcp   自分のIPのみ
  80/tcp   0.0.0.0/0, ::/0
  443/tcp  0.0.0.0/0, ::/0

Outbound
  all allowed
```

MySQLの3306、backendの8080、frontendの8000は外部公開しない。Caddy経由でだけ公開する。

## 8. Elastic IPを割り当て

EC2のパブリックIPは停止/開始で変わる可能性があるため、Elastic IPを割り当てる。

1. EC2 > Elastic IPs
2. Allocate Elastic IP
3. Associate Elastic IP address
4. 作成したEC2へ紐付け

使っていないElastic IPは課金対象になるので、必ずEC2へ紐付ける。

## 9. Route 53でDNS設定

Hosted zoneでAレコードを作る。

```text
example.com       A  Elastic IP
www.example.com   A  Elastic IP
api.example.com   A  Elastic IP
admin.example.com A  Elastic IP
```

最初は `example.com` と `api.example.com` だけでもよい。

## 10. EC2へSSH接続

ローカルPCから接続する。

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@YOUR_ELASTIC_IP
```

接続後、更新する。

```bash
sudo apt update
sudo apt -y upgrade
sudo timedatectl set-timezone Asia/Tokyo
```

## 11. 基本パッケージ導入

```bash
sudo apt install -y git curl unzip build-essential mysql-server awscli
```

Goを入れる。バージョンは開発環境に合わせて `1.25.x` 系を使う。

```bash
cd /tmp
curl -LO https://go.dev/dl/go1.25.1.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.1.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```

`t3.micro` などx86_64を使う場合は `linux-amd64` のtarを使う。

Denoを入れる。

```bash
curl -fsSL https://deno.land/install.sh | sh
echo 'export DENO_INSTALL="$HOME/.deno"' >> ~/.profile
echo 'export PATH="$DENO_INSTALL/bin:$PATH"' >> ~/.profile
source ~/.profile
deno --version
```

## 12. アプリ実行ユーザー作成

```bash
sudo useradd --system --create-home --shell /bin/bash yutagame
sudo mkdir -p /opt/yutagame
sudo chown -R yutagame:yutagame /opt/yutagame
```

## 13. MySQL初期設定

```bash
sudo systemctl enable mysql
sudo systemctl start mysql
sudo mysql_secure_installation
```

DBとユーザーを作る。

```bash
sudo mysql
```

MySQL内で実行:

```sql
CREATE DATABASE yutagame CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER 'yutagame_app'@'localhost' IDENTIFIED BY 'CHANGE_STRONG_PASSWORD';
GRANT ALL PRIVILEGES ON yutagame.* TO 'yutagame_app'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

複数サイトを同じEC2で動かす場合は、サイトごとにDB名とDBユーザーを分ける。

```text
yutagame
another_site
```

## 14. 環境変数ファイル作成

GitHubには機密情報を置かない。EC2上に置く。

```bash
sudo mkdir -p /etc/yutagame
sudo touch /etc/yutagame/backend.env /etc/yutagame/frontend.env
sudo chown -R yutagame:yutagame /etc/yutagame
sudo chmod 700 /etc/yutagame
sudo chmod 600 /etc/yutagame/*.env
```

backend:

```bash
sudo -u yutagame nano /etc/yutagame/backend.env
```

例:

```env
DB_USER=yutagame_app
DB_PASSWORD=CHANGE_STRONG_PASSWORD
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=yutagame
BACKEND_PORT=8080
JWT_SECRET=CHANGE_LONG_RANDOM_SECRET

DEFAULT_ADMIN_EMAIL=admin@example.com
DEFAULT_ADMIN_PASSWORD_HASH=$2a$10$53xmr3o2m1Tuxo0IQFfko.Suq4Z426P16q4eStnZ9u4acD0WlyeAq
DEFAULT_ADMIN_NAME=デフォルト管理者アカウント
DEFAULT_ADMIN_ROLE=ADMIN

SMTP_HOST=email-smtp.ap-northeast-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=CHANGE_SES_SMTP_USERNAME
SMTP_PASSWORD=CHANGE_SES_SMTP_PASSWORD
SMTP_FROM=no-reply@example.com
SMTP_USE_TLS=false
CONTACT_NOTIFY_TO=admin@example.com
ADMIN_SITE_URL=https://example.com/admin
RUN_AUTO_MIGRATE=false
MIGRATIONS_DIR=migrations
```

frontend:

```bash
sudo -u yutagame nano /etc/yutagame/frontend.env
```

例:

```env
APP_BASE_URL=https://api.example.com/api
ADMIN_BASE_URL=https://api.example.com/api
SITE_ORIGIN=https://example.com
PUBLIC_SITE_ORIGIN=https://example.com
FRONTEND_LOG_DIR=/var/log/yutagame/frontend
```

FreshのIslandなどブラウザ側で動くJavaScriptは、サーバー側の `Deno.env.get()`
を直接読めない。このプロジェクトでは `_app.tsx` が `APP_BASE_URL` /
`ADMIN_BASE_URL` / `SITE_ORIGIN` を公開runtime
configとしてHTMLへ埋め込み、ブラウザ側のAPIクライアントがそれを読む。

環境変数はMarkdown形式ではなく、生のURLを書く。

```env
ADMIN_BASE_URL=https://api.package-forest.to-force.com/api
```

次のように書くとURLとして解釈できない。

```env
ADMIN_BASE_URL=[https://api.package-forest.to-force.com/api](https://api.package-forest.to-force.com/api)
```

値を変更したらfrontendを再起動する。

```bash
sudo systemctl restart yutagame-frontend
```

## 15. ソースコード配置

まずはEC2上でgit cloneする。

```bash
sudo -u yutagame git clone https://github.com/YOUR_ACCOUNT/YOUR_REPO.git /opt/yutagame/app
```

ディレクトリ例:

```text
/opt/yutagame/app/yutagame-backend
/opt/yutagame/app/yutagame-frontend
```

プライベートリポジトリの場合は、Deploy keyかGitHub fine-grained
tokenを使う。tokenをshell historyに残さないよう注意する。

## 16. ログ/ストレージディレクトリ

```bash
sudo mkdir -p /var/log/yutagame/backend /var/log/yutagame/frontend
sudo mkdir -p /var/lib/yutagame/storage
sudo chown -R yutagame:yutagame /var/log/yutagame /var/lib/yutagame
```

現状のバックエンドは作業ディレクトリ配下の `storage/`
を見るため、まずはアプリ配下へシンボリックリンクを作る。

```bash
sudo -u yutagame mkdir -p /opt/yutagame/app/yutagame-backend
sudo -u yutagame ln -sfn /var/lib/yutagame/storage /opt/yutagame/app/yutagame-backend/storage
```

## 17. ビルド

backend:

```bash
cd /opt/yutagame/app/yutagame-backend
sudo -u yutagame /usr/local/go/bin/go build -o yutagame-backend .
sudo -u yutagame /usr/local/go/bin/go build -o yutagame-migrate ./cmd/migrate
```

frontend:

```bash
cd /opt/yutagame/app/yutagame-frontend
sudo -u yutagame /home/yutagame/.deno/bin/deno task build
```

必要に応じて `deno cache` も実行する。

## 18. 初回マイグレーション

空DBにテーブルを作り、デフォルト管理者だけ投入する。

```bash
cd /opt/yutagame/app/yutagame-backend
sudo -u yutagame bash -lc 'set -a; source /etc/yutagame/backend.env; set +a; ./yutagame-migrate up'
```

このコマンドは `migrations/` 配下のSQLをgooseで実行し、DB内の `goose_db_version`
に適用履歴を記録する。`up` 成功後に、環境変数 `DEFAULT_ADMIN_EMAIL`
などを使ってデフォルト管理者を1件だけ作成する。

確認:

```bash
mysql -u yutagame_app -p yutagame -e "SHOW TABLES LIKE 'goose_db_version'; SELECT version_id,is_applied FROM goose_db_version; SELECT id,email,name,role_type FROM admins;"
```

本番ではローカルの `init/` seedは使わない。

## 19. systemd設定

backend service:

```bash
sudo nano /etc/systemd/system/yutagame-backend.service
```

```ini
[Unit]
Description=Yutagame Backend
After=network.target mysql.service

[Service]
WorkingDirectory=/opt/yutagame/app/yutagame-backend
EnvironmentFile=/etc/yutagame/backend.env
ExecStart=/opt/yutagame/app/yutagame-backend/yutagame-backend
Restart=always
RestartSec=5
User=yutagame
Group=yutagame

[Install]
WantedBy=multi-user.target
```

frontend service:

```bash
sudo nano /etc/systemd/system/yutagame-frontend.service
```

```ini
[Unit]
Description=Yutagame Frontend
After=network.target yutagame-backend.service

[Service]
WorkingDirectory=/opt/yutagame/app/yutagame-frontend
EnvironmentFile=/etc/yutagame/frontend.env
ExecStart=/home/yutagame/.deno/bin/deno run -A main.ts
Restart=always
RestartSec=5
User=yutagame
Group=yutagame

[Install]
WantedBy=multi-user.target
```

起動:

```bash
sudo systemctl daemon-reload
sudo systemctl enable yutagame-backend yutagame-frontend
sudo systemctl start yutagame-backend yutagame-frontend
sudo systemctl status yutagame-backend
sudo systemctl status yutagame-frontend
```

ログ確認:

```bash
journalctl -u yutagame-backend -f
journalctl -u yutagame-frontend -f
```

## 20. CaddyでHTTPS公開

Caddyを入れる。

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install -y caddy
```

Caddyfile:

```bash
sudo nano /etc/caddy/Caddyfile
```

例:

```caddyfile
example.com, www.example.com {
  reverse_proxy 127.0.0.1:8000
}

api.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

反映:

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
sudo systemctl status caddy
```

DNSが向いていて、80/443が開いていればLet's Encrypt証明書は自動取得される。

## 21. Amazon SESでメール送信

問い合わせ/おすすめゲーム投稿通知用。

1. Amazon SESを開く
2. 東京リージョンを選ぶ
3. Verified identitiesで送信元ドメインまたはメールアドレスを検証
4. Route 53にDKIMレコードを追加
5. SMTP credentialsを作る
6. `/etc/yutagame/backend.env` に設定
7. サンドボックス解除申請をする

設定例:

```env
SMTP_HOST=email-smtp.ap-northeast-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=SES_SMTP_USERNAME
SMTP_PASSWORD=SES_SMTP_PASSWORD
SMTP_FROM=no-reply@example.com
CONTACT_NOTIFY_TO=admin@example.com
```

反映:

```bash
sudo systemctl restart yutagame-backend
```

## 22. DBバックアップ

バックアップスクリプトを作る。

```bash
sudo nano /usr/local/bin/yutagame-db-backup.sh
```

```bash
#!/usr/bin/env bash
set -euo pipefail

set -a
source /etc/yutagame/backend.env
set +a

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/var/backups/yutagame
S3_BUCKET=s3://yutagame-prod-backup-xxxxxxxx/yutagame

mkdir -p "$BACKUP_DIR"
mysqldump -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" | gzip > "$BACKUP_DIR/yutagame_$DATE.sql.gz"
aws s3 cp "$BACKUP_DIR/yutagame_$DATE.sql.gz" "$S3_BUCKET/"
find "$BACKUP_DIR" -type f -name 'yutagame_*.sql.gz' -mtime +7 -delete
```

権限:

```bash
sudo chmod 700 /usr/local/bin/yutagame-db-backup.sh
sudo chown root:root /usr/local/bin/yutagame-db-backup.sh
```

手動テスト:

```bash
sudo /usr/local/bin/yutagame-db-backup.sh
aws s3 ls s3://yutagame-prod-backup-xxxxxxxx/yutagame/
```

cron:

```bash
sudo crontab -e
```

```cron
15 3 * * * /usr/local/bin/yutagame-db-backup.sh >> /var/log/yutagame/db-backup.log 2>&1
```

最低月1回は復元テストをする。

## 23. CloudWatch Agentで監視

CloudWatch Agentを入れる。

```bash
wget https://amazoncloudwatch-agent-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/ubuntu/arm64/latest/amazon-cloudwatch-agent.deb -O /tmp/amazon-cloudwatch-agent.deb
sudo dpkg -i /tmp/amazon-cloudwatch-agent.deb
```

x86_64の場合は `amd64/latest` を使う。

設定例:

```bash
sudo nano /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json
```

```json
{
    "metrics": {
        "append_dimensions": {
            "InstanceId": "${aws:InstanceId}"
        },
        "metrics_collected": {
            "mem": {
                "measurement": ["mem_used_percent"],
                "metrics_collection_interval": 60
            },
            "disk": {
                "measurement": ["used_percent"],
                "resources": ["/"],
                "metrics_collection_interval": 60
            }
        }
    }
}
```

起動:

```bash
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json \
  -s
```

CloudWatch Alarm:

- CPUUtilization > 80%
- mem_used_percent > 80%
- disk used_percent > 80%

## 24. デプロイ方法

最初は手動デプロイでよい。

```bash
cd /opt/yutagame/app
sudo -u yutagame git pull

cd /opt/yutagame/app/yutagame-backend
sudo -u yutagame /usr/local/go/bin/go build -o yutagame-backend .
sudo -u yutagame /usr/local/go/bin/go build -o yutagame-migrate ./cmd/migrate
sudo -u yutagame bash -lc 'set -a; source /etc/yutagame/backend.env; set +a; ./yutagame-migrate up'

cd /opt/yutagame/app/yutagame-frontend
sudo -u yutagame /home/yutagame/.deno/bin/deno task build

sudo systemctl restart yutagame-backend yutagame-frontend
```

慣れてきたらGitHub Actionsにする。

GitHub Secretsに入れるもの:

- `EC2_HOST`
- `EC2_USER`
- `EC2_SSH_KEY`

SecretsにはDBパスワードを入れない。DBパスワードなどはEC2上の
`/etc/yutagame/*.env` に置く。

GitHub Actions例:

```yaml
name: Deploy Production

on:
    workflow_dispatch:
    push:
        branches: [main]

jobs:
    deploy:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
            - name: Deploy by SSH
              uses: appleboy/ssh-action@v1.0.3
              with:
                  host: ${{ secrets.EC2_HOST }}
                  username: ${{ secrets.EC2_USER }}
                  key: ${{ secrets.EC2_SSH_KEY }}
                  script: |
                      set -e
                      cd /opt/yutagame/app
                      sudo -u yutagame git pull
                      cd /opt/yutagame/app/yutagame-backend
                      sudo -u yutagame /usr/local/go/bin/go build -o yutagame-backend .
                      sudo -u yutagame /usr/local/go/bin/go build -o yutagame-migrate ./cmd/migrate
                      sudo -u yutagame bash -lc 'set -a; source /etc/yutagame/backend.env; set +a; ./yutagame-migrate up'
                      cd /opt/yutagame/app/yutagame-frontend
                      sudo -u yutagame /home/yutagame/.deno/bin/deno task build
                      sudo systemctl restart yutagame-backend yutagame-frontend
```

## 25. OpenTofuで管理する範囲

最初から全部IaCにすると難しい。最初は以下だけOpenTofu化する。

- Security Group
- EC2
- Elastic IP
- Route 53 record
- S3 backup bucket
- IAM role

MySQL内部のユーザー作成、アプリ配置、systemdは最初は手順書運用でよい。慣れたらAnsibleなどへ移す。

## 26. 複数Webサイトを同じEC2で運用する場合

サイトごとに分けるもの:

- Linuxユーザー
- `/opt/{site-name}`
- `/etc/{site-name}/*.env`
- `/var/log/{site-name}`
- MySQL DBスキーマ
- systemd service名
- Caddyのドメイン設定

例:

```text
/opt/yutagame
/opt/another-site

/etc/yutagame/backend.env
/etc/another-site/backend.env

DB: yutagame
DB: another_site
```

負荷が上がったら別EC2へ分離する。

## 27. マイグレーション管理

本番環境ではgooseでSQL migrationを管理する。適用履歴はDB内の `goose_db_version`
に保存される。

初期migrationは
`migrations/000001_initial_schema.sql`。今後DBスキーマを変更するときは、必ず次の番号のmigrationファイルを追加する。

```text
migrations/
  000001_initial_schema.sql
  000002_add_xxx.sql
  000003_change_yyy.sql
```

デプロイ時は、アプリ再起動前に必ず以下を実行する。

```bash
./yutagame-migrate up
```

`down`
は直近のmigrationを戻すためのコマンドだが、本番ではバックアップ取得後に慎重に使う。

```bash
./yutagame-migrate status
./yutagame-migrate down
```

特に以下の変更は、AutoMigrateに頼らずSQL migrationとして明示する。

- カラム名変更
- カラム削除
- データ移行
- インデックスの大きな変更
- ダウンタイムを避けたい変更

## 28. 公開前チェックリスト

- AWS root userにMFAがある
- 予算アラートがある
- EC2 Security Groupで3306/8000/8080を公開していない
- Route 53 AレコードがElastic IPを向いている
- CaddyでHTTPSが有効
- `https://example.com` が表示できる
- `https://api.example.com/api/site-stats` が返る
- `/admin` がnoindexになっている
- MySQLバックアップがS3へ保存される
- CloudWatchでCPU/メモリ/ディスクが見える
- SESで通知メールが届く
- `/etc/yutagame/*.env` の権限が `600`
- `goose_db_version` が存在し、version 1 が適用済み
- デフォルト管理者でログインできる
- ログイン後、初期パスワードを変更する

## 29. 最初にやる順番

1. AWSアカウント作成、MFA、予算アラート
2. Route 53でドメイン取得
3. S3バックアップバケット作成
4. EC2 Role作成
5. EC2作成、Elastic IP付与
6. DNS設定
7. EC2へSSH
8. MySQL/Go/Deno/Caddy導入
9. DB/DBユーザー作成
10. `/etc/yutagame/*.env` 作成
11. ソース配置、ビルド
12. `./yutagame-migrate up` で初回DB作成
13. systemd起動
14. CaddyでHTTPS公開
15. SES設定
16. S3バックアップ設定
17. CloudWatch監視設定
18. GitHub Actionsデプロイ設定
