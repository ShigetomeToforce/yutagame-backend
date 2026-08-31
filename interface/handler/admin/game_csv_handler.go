package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GameCSVExportRequest struct {
	Columns  []string `json:"columns"`
	Encoding string   `json:"encoding"`
}

type GameCSVApplyRequest struct {
	Operations []usecaseAdmin.GameCSVOperation `json:"operations"`
}

func (h *GameHandler) ExportCSV(c echo.Context) error {
	var req GameCSVExportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	encoding := normalizeCSVEncoding(req.Encoding)
	selected := normalizeGameColumns(req.Columns)

	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetAllGames(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	headers := []string{"id"}
	headers = append(headers, selected...)

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write(headers); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	for _, item := range items {
		row := make([]string, 0, len(headers))
		row = append(row, strconv.FormatInt(item.ID, 10))
		for _, col := range selected {
			switch col {
			case "name":
				row = append(row, item.Name)
			case "kana":
				row = append(row, item.Kana)
			case "overview":
				row = append(row, item.Overview)
			case "code":
				row = append(row, item.Code)
			case "manufacturer_id":
				row = append(row, strconv.FormatInt(item.ManufacturerID, 10))
			case "machine_id":
				row = append(row, strconv.FormatInt(item.MachineID, 10))
			case "genre_id":
				row = append(row, strconv.FormatInt(item.GenreID, 10))
			case "sub_genre":
				row = append(row, item.SubGenre)
			case "catch_copy":
				row = append(row, item.CatchCopy)
			case "sub_catch":
				row = append(row, item.SubCatch)
			case "list_price":
				row = append(row, strconv.FormatInt(int64(item.ListPrice), 10))
			case "release_date":
				row = append(row, item.ReleaseDate.Format("2006-01-02"))
			case "official_site_url":
				row = append(row, item.OfficialSiteURL)
			case "youtube_url":
				row = append(row, item.YouTubeURL)
			case "is_play":
				row = append(row, strconv.FormatBool(item.IsPlay))
			case "is_clear":
				row = append(row, strconv.FormatBool(item.IsClear))
			case "is_favourite":
				row = append(row, strconv.FormatBool(item.IsFavourite))
			}
		}
		if err := writer.Write(row); err != nil {
			return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	payload, err := convertCSVEncoding(buf.Bytes(), encoding)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	filename := fmt.Sprintf("games_%s.csv", time.Now().Format("20060102_150405"))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "text/csv", payload)
}

func (h *GameHandler) PreviewImportCSV(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "CSVファイルを選択してください。"})
	}

	encoding := normalizeCSVEncoding(c.FormValue("encoding"))
	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "CSVファイルを開けませんでした。"})
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "CSVファイルの読み込みに失敗しました。"})
	}

	text, err := decodeCSVBytes(raw, encoding)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	rows, err := parseGameCSVRows(text)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	preview, err := h.gameUseCase.BuildCSVPreview(ctx, rows)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, preview)
}

func (h *GameHandler) ApplyImportCSV(c echo.Context) error {
	var req GameCSVApplyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if len(req.Operations) == 0 {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "更新対象が選択されていません。"})
	}

	ctx := c.Request().Context()
	created, updated, err := h.gameUseCase.ApplyCSVOperations(ctx, req.Operations)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"createdCount": created,
		"updatedCount": updated,
		"message":      "CSVインポートが完了しました。",
	})
}

func normalizeGameColumns(columns []string) []string {
	allow := map[string]bool{
		"name":              true,
		"kana":              true,
		"overview":          true,
		"code":              true,
		"manufacturer_id":   true,
		"machine_id":        true,
		"genre_id":          true,
		"sub_genre":         true,
		"catch_copy":        true,
		"sub_catch":         true,
		"list_price":        true,
		"release_date":      true,
		"official_site_url": true,
		"youtube_url":       true,
		"is_play":           true,
		"is_clear":          true,
		"is_favourite":      true,
	}
	seen := map[string]bool{}
	result := make([]string, 0, 17)
	for _, col := range columns {
		normalized := strings.ToLower(strings.TrimSpace(col))
		if allow[normalized] && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	if len(result) == 0 {
		return []string{"name", "kana", "overview", "code", "manufacturer_id", "machine_id", "genre_id", "sub_genre", "catch_copy", "sub_catch", "list_price", "release_date", "official_site_url", "youtube_url", "is_play", "is_clear", "is_favourite"}
	}
	return result
}

func parseGameCSVRows(csvText string) ([]usecaseAdmin.GameCSVRowInput, error) {
	reader := csv.NewReader(strings.NewReader(csvText))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV解析に失敗しました: %w", err)
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("CSVにヘッダー行がありません。")
	}

	headers := make(map[int]string, len(records[0]))
	hasID := false
	for i, raw := range records[0] {
		header := normalizeGameCSVHeader(raw)
		headers[i] = header
		if header == "id" {
			hasID = true
		}
	}
	if !hasID {
		return nil, fmt.Errorf("id 列が必須です。")
	}

	rows := make([]usecaseAdmin.GameCSVRowInput, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		row := usecaseAdmin.GameCSVRowInput{RowNumber: int64(i + 1)}
		for colIndex, value := range record {
			header, ok := headers[colIndex]
			if !ok {
				continue
			}
			trimmed := strings.TrimSpace(value)
			switch header {
			case "id":
				if trimmed == "" {
					continue
				}
				id, parseErr := strconv.ParseInt(trimmed, 10, 64)
				if parseErr != nil {
					invalid := int64(-1)
					row.ID = &invalid
					continue
				}
				row.ID = &id
			case "name":
				v := value
				row.Name = &v
			case "kana":
				v := value
				row.Kana = &v
			case "overview":
				v := value
				row.Overview = &v
			case "code":
				v := value
				row.Code = &v
			case "manufacturer_id":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 64)
				if parseErr != nil {
					continue
				}
				row.ManufacturerID = &num
			case "machine_id":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 64)
				if parseErr != nil {
					continue
				}
				row.MachineID = &num
			case "genre_id":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 64)
				if parseErr != nil {
					continue
				}
				row.GenreID = &num
			case "sub_genre":
				v := value
				row.SubGenre = &v
			case "catch_copy":
				v := value
				row.CatchCopy = &v
			case "sub_catch":
				v := value
				row.SubCatch = &v
			case "list_price":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 32)
				if parseErr != nil {
					continue
				}
				v := int32(num)
				row.ListPrice = &v
			case "release_date":
				v := trimmed
				row.ReleaseDate = &v
			case "official_site_url":
				v := value
				row.OfficialSiteURL = &v
			case "youtube_url":
				v := value
				row.YouTubeURL = &v
			case "is_play":
				parsed, ok := parseCSVBool(trimmed)
				if ok {
					row.IsPlay = &parsed
				}
			case "is_clear":
				parsed, ok := parseCSVBool(trimmed)
				if ok {
					row.IsClear = &parsed
				}
			case "is_favourite":
				parsed, ok := parseCSVBool(trimmed)
				if ok {
					row.IsFavourite = &parsed
				}
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func parseCSVBool(raw string) (bool, bool) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "true", "1", "yes", "y", "on":
		return true, true
	case "false", "0", "no", "n", "off":
		return false, true
	default:
		return false, false
	}
}

func normalizeGameCSVHeader(raw string) string {
	h := strings.ToLower(strings.TrimSpace(raw))
	switch h {
	case "id", "識別子":
		return "id"
	case "name", "名前":
		return "name"
	case "kana", "カナ":
		return "kana"
	case "overview", "概要":
		return "overview"
	case "code", "コード":
		return "code"
	case "manufacturer_id", "manufacturerid", "メーカーid":
		return "manufacturer_id"
	case "machine_id", "machineid", "機種id":
		return "machine_id"
	case "genre_id", "genreid", "ジャンルid":
		return "genre_id"
	case "sub_genre", "subgenre", "サブジャンル":
		return "sub_genre"
	case "catch_copy", "catchcopy", "キャッチコピー":
		return "catch_copy"
	case "sub_catch", "subcatch", "サブキャッチ":
		return "sub_catch"
	case "list_price", "listprice", "価格":
		return "list_price"
	case "release_date", "releasedate", "発売日":
		return "release_date"
	case "official_site_url", "officialsiteurl", "公式url":
		return "official_site_url"
	case "youtube_url", "youtubeurl", "youtube":
		return "youtube_url"
	case "is_play", "isplay", "プレイ済み":
		return "is_play"
	case "is_clear", "isclear", "クリア済み":
		return "is_clear"
	case "is_favourite", "isfavourite", "is_favorite", "isfavorite", "お気に入り":
		return "is_favourite"
	default:
		return ""
	}
}
