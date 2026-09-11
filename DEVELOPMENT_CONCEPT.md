# YUTAGAME Backend Development Concept

## 1. 目的

YUTAGAMEバックエンドは、管理画面向けの更新系APIと公開画面向けの参照系APIを明確に分離し、次の3点を継続的に満たすことを目的とする。

- 安全に更新できること
- 仕様変更時に影響範囲を追いやすいこと
- 運用機能（CSV、画像）を同じ設計原則で増やせること

## 2. 全体構造

### 2.1 レイヤー責務

- domain/model
  - DBスキーマとの対応モデル、リレーション定義。
- infrastructure/database
  - GORMを使った永続化。Find/Create/Update/Deleteや、CSV向け補助クエリを配置。
- application/usecase
  - ビジネスルール本体。重複チェック、差分判定、更新適用フローを管理。
- interface/handler
  - HTTP入出力、バインド、レスポンスコード制御。

### 2.2 どこでルーティングしているか

ルーティングは [yutagame-backend/main.go](main.go) で一元管理している。

- 公開APIグループ: /api/app
- 管理APIグループ: /api/admin
- 例外公開: /api/admin/login のみ
- 認証適用: /api/admin グループに AdminGuard を適用

認証ミドルウェア本体は
[yutagame-backend/interface/middleware/admin_guard.go](interface/middleware/admin_guard.go)。

## 3. DIの考え方と実装場所

### 3.1 DIの考え方

依存関係の向きは次の順に固定する。

- Handler -> UseCase -> Repository -> DB

この向きを守ることで、HTTP仕様変更と業務仕様変更を分離しやすくする。

### 3.2 DIをどこでやっているか

DIは [yutagame-backend/main.go](main.go) で行う。

- DB接続初期化
- Repositoryインスタンス生成
- UseCaseインスタンス生成
- Handlerインスタンス生成
- ルートにHandlerを割り当て

この方針により、各層のファイルは依存注入コードを持たず、責務を保てる。

## 4. 共通処理をどこでまとめるか

### 4.1 Handler共通

共通レスポンス型や配列パラメータ変換は
[yutagame-backend/interface/handler](interface/handler) 配下で管理。

- response.go: 共通レスポンス構造
- pagenation.go: ページングレスポンス
- util.go: クエリ変換など

### 4.2 UseCase共通

ページング検索共通は
[yutagame-backend/application/usecase/pagination_util.go](application/usecase/pagination_util.go)。

コード重複チェックは
[yutagame-backend/application/usecase/admin/code_validation.go](application/usecase/admin/code_validation.go)。

### 4.3 Repository共通

一覧取得と件数取得の共通処理は
[yutagame-backend/infrastructure/database/common_query.go](infrastructure/database/common_query.go)。

## 5. ファイルダウンロード機能はどこで、どう実装しているか

### 5.1 実装場所

CSVエクスポートは各リソースのCSVハンドラに実装している。

- [yutagame-backend/interface/handler/admin/genre_csv_handler.go](interface/handler/admin/genre_csv_handler.go)
- [yutagame-backend/interface/handler/admin/game_csv_handler.go](interface/handler/admin/game_csv_handler.go)
- [yutagame-backend/interface/handler/admin/machine_csv_handler.go](interface/handler/admin/machine_csv_handler.go)
- [yutagame-backend/interface/handler/admin/manufacturer_csv_handler.go](interface/handler/admin/manufacturer_csv_handler.go)
- [yutagame-backend/interface/handler/admin/keyword_csv_handler.go](interface/handler/admin/keyword_csv_handler.go)

### 5.2 処理フロー

- リクエストの列指定と文字コードを受け取る
- UseCase経由で全件取得
- encoding/csv でCSV生成
- 文字コード変換
  - UTF-8
  - BOM付きUTF-8
  - Shift-JIS
- Content-Disposition を設定して c.Blob で返却

### 5.3 文字コード変換の共通化

実質的な共通関数は
genre側ハンドラに配置され、他CSVハンドラから同一パッケージ内関数として利用している。

- normalizeCSVEncoding
- convertCSVEncoding
- decodeCSVBytes

今後さらに明確化する場合は csv_common_handler.go へ切り出す。

## 6. CSVインポート機能の設計

### 6.1 3段階API

- export
- import/preview
- import/apply

### 6.2 previewの責務

UseCaseで create / update / skip を判定し、UIがそのまま表示できる構造を返す。

- Operation
  - action
  - selectable
  - diffs
  - reason

### 6.3 applyの責務

選択された operation のみ適用する。

- create
  - 必須列確認
  - code重複確認
  - 新規作成
- update
  - 差分再計算
  - 必要列のみ部分更新

### 6.4 誤差分対策

改行コード差や前後空白で誤検知しないよう、文字列比較時に正規化関数を通す。

## 7. 画像アップロード設計

画像処理は各Handlerにあり、保存共通関数は
[yutagame-backend/interface/handler/admin/image_util.go](interface/handler/admin/image_util.go)
に置く。

- POST /:id/image で保存
- DELETE /:id/image で削除
- DBには image_key を保存
- 配信は e.Static("/images", "storage") で提供

## 8. 開発時の実践ルール

- ルート追加時は main.go の保護境界を必ず確認する
- UseCaseに業務判断を寄せ、Handlerを薄く保つ
- Saveで副作用が出る場合は UpdateFieldsByID を使う
- JSONのnullアクセスでフロントが落ちないよう、空配列初期化を徹底する
- コメントはコードの逐語訳ではなく、JST境界、重複防止、互換性など「なぜその実装か」を説明する
- 一覧APIでは全件取得後のGo側絞り込みを避け、WHERE・ORDER・LIMITをDBへ任せる
- 複数日の集計は日数分ループせず、GROUP BYで一括取得する

## 9. 日常コマンド

- gofmt -w ./...
- go test ./...
- docker compose up -d --build
- docker compose restart backend

Swagger更新:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go --output docs
```

## 10. テスト方針

`*_test.go`
は対象コードと同じpackage付近に置き、まず外部DBなしで再現できる業務ルールを高速に検証する。

現在の主な単体テスト:

- ページングの境界値と不正limit
- PV訪問者識別、パス正規化、UTF-8切り詰め
- アクセス分析の日次・月次範囲とJST
- お知らせ公開日時のJST解釈
- 公開ゲーム検索のsortとpage/limit補正

今後の優先順位:

1. MySQLを使ったRepository統合テスト（検索条件、日次重複、一括集計）
2. UseCaseのRepositoryをinterface化し、問い合わせ・お気に入り・ランキングをmockテスト
3. Echo HandlerのHTTPステータスとJSON契約テスト

外部依存を伴うテストを増やす際も、通常の `go test ./...`
で実行できる単体テストと分離する。

## 11. アクセス集計

サイト全体のPV/UUは `page_view_logs` を正とする。

- `routes/_middleware.ts` が公開HTMLページの成功したGETを記録する
- `PageViewLogRepository.CreateDaily` がDB一意制約を利用して二重登録を無視する
- 日付はコンテナのローカル設定ではなく `Asia/Tokyo` を明示する
- 日次表はrepositoryで日付GROUP BYし、日数分のクエリを発行しない

`game_view_logs`
はゲーム詳細の人気ランキング専用であり、サイト全体PVには使わない。

## 12. 拡張方針

- CSV列定義の共通テーブル化
- apply処理のトランザクション境界強化
- 行単位の詳細エラー返却の標準化
