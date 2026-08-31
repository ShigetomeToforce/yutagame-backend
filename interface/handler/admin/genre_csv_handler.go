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
	"golang.org/x/text/encoding/japanese"
)

type GenreCSVExportRequest struct {
	Columns  []string `json:"columns"`
	Encoding string   `json:"encoding"`
}

type GenreCSVApplyRequest struct {
	Operations []usecaseAdmin.GenreCSVOperation `json:"operations"`
}

func (h *GenreHandler) ExportCSV(c echo.Context) error {
	var req GenreCSVExportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	encoding := normalizeCSVEncoding(req.Encoding)
	selected := normalizeGenreColumns(req.Columns)

	ctx := c.Request().Context()
	genres, err := h.genreUseCase.GetAllGenres(ctx)
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

	for _, g := range genres {
		row := make([]string, 0, len(headers))
		row = append(row, strconv.FormatInt(g.ID, 10))

		for _, col := range selected {
			switch col {
			case "name":
				row = append(row, g.Name)
			case "kana":
				row = append(row, g.Kana)
			case "overview":
				row = append(row, g.Overview)
			case "code":
				row = append(row, g.Code)
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

	filename := fmt.Sprintf("genres_%s.csv", time.Now().Format("20060102_150405"))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Blob(http.StatusOK, "text/csv", payload)
}

func (h *GenreHandler) PreviewImportCSV(c echo.Context) error {
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

	rows, err := parseGenreCSVRows(text)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	preview, err := h.genreUseCase.BuildCSVPreview(ctx, rows)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, preview)
}

func (h *GenreHandler) ApplyImportCSV(c echo.Context) error {
	var req GenreCSVApplyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if len(req.Operations) == 0 {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "更新対象が選択されていません。"})
	}

	ctx := c.Request().Context()
	created, updated, err := h.genreUseCase.ApplyCSVOperations(ctx, req.Operations)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"createdCount": created,
		"updatedCount": updated,
		"message":      "CSVインポートが完了しました。",
	})
}

func normalizeCSVEncoding(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "utf-8", "utf8":
		return "utf8"
	case "utf-8-bom", "utf8bom", "bom-utf8":
		return "utf8bom"
	case "shift-jis", "shift_jis", "sjis", "cp932":
		return "shift_jis"
	default:
		return "utf8"
	}
}

func convertCSVEncoding(utf8Payload []byte, encoding string) ([]byte, error) {
	switch encoding {
	case "utf8":
		return utf8Payload, nil
	case "utf8bom":
		return append([]byte{0xEF, 0xBB, 0xBF}, utf8Payload...), nil
	case "shift_jis":
		encoded, err := japanese.ShiftJIS.NewEncoder().Bytes(utf8Payload)
		if err != nil {
			return nil, fmt.Errorf("SHIFT-JISへの変換に失敗しました: %w", err)
		}
		return encoded, nil
	default:
		return nil, fmt.Errorf("未対応の文字コードです: %s", encoding)
	}
}

func decodeCSVBytes(raw []byte, encoding string) (string, error) {
	trimmed := raw
	if len(trimmed) >= 3 && trimmed[0] == 0xEF && trimmed[1] == 0xBB && trimmed[2] == 0xBF {
		trimmed = trimmed[3:]
	}

	switch encoding {
	case "utf8", "utf8bom":
		return string(trimmed), nil
	case "shift_jis":
		decoded, err := japanese.ShiftJIS.NewDecoder().Bytes(trimmed)
		if err != nil {
			return "", fmt.Errorf("SHIFT-JIS CSVのデコードに失敗しました: %w", err)
		}
		return string(decoded), nil
	default:
		return "", fmt.Errorf("未対応の文字コードです: %s", encoding)
	}
}

func normalizeGenreColumns(columns []string) []string {
	allow := map[string]bool{
		"name":     true,
		"kana":     true,
		"overview": true,
		"code":     true,
	}
	seen := map[string]bool{}
	result := make([]string, 0, 4)
	for _, col := range columns {
		normalized := strings.ToLower(strings.TrimSpace(col))
		if allow[normalized] && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	if len(result) == 0 {
		return []string{"name", "kana", "overview", "code"}
	}
	return result
}

func parseGenreCSVRows(csvText string) ([]usecaseAdmin.GenreCSVRowInput, error) {
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
		header := normalizeCSVHeader(raw)
		headers[i] = header
		if header == "id" {
			hasID = true
		}
	}
	if !hasID {
		return nil, fmt.Errorf("id 列が必須です。")
	}

	rows := make([]usecaseAdmin.GenreCSVRowInput, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		row := usecaseAdmin.GenreCSVRowInput{RowNumber: int64(i + 1)}

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
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func normalizeCSVHeader(raw string) string {
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
	default:
		return ""
	}
}
