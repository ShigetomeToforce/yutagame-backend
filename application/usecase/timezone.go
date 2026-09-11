package usecase

import "time"

// JapanLocation はOSやコンテナのタイムゾーン設定に依存しない業務上の日付境界です。
var JapanLocation = time.FixedZone("Asia/Tokyo", 9*60*60)
