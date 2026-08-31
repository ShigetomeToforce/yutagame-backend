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

type MachineCSVExportRequest struct {
	Columns  []string `json:"columns"`
	Encoding string   `json:"encoding"`
}

type MachineCSVApplyRequest struct {
	Operations []usecaseAdmin.MachineCSVOperation `json:"operations"`
}

func (h *MachineHandler) ExportCSV(c echo.Context) error {
	var req MachineCSVExportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	encoding := normalizeCSVEncoding(req.Encoding)
	selected := normalizeMachineColumns(req.Columns)

	ctx := c.Request().Context()
	items, err := h.machineUseCase.GetAllMachines(ctx)
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
			case "abbreviation":
				row = append(row, item.Abbreviation)
			case "manufacturer_id":
				row = append(row, strconv.FormatInt(item.ManufacturerID, 10))
			case "machine_type":
				row = append(row, item.MachineType)
			case "release_date":
				row = append(row, item.ReleaseDate.Format("2006-01-02"))
			case "sort_order":
				row = append(row, strconv.FormatInt(int64(item.SortOrder), 10))
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

	filename := fmt.Sprintf("machines_%s.csv", time.Now().Format("20060102_150405"))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "text/csv", payload)
}

func (h *MachineHandler) PreviewImportCSV(c echo.Context) error {
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

	rows, err := parseMachineCSVRows(text)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	preview, err := h.machineUseCase.BuildCSVPreview(ctx, rows)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, preview)
}

func (h *MachineHandler) ApplyImportCSV(c echo.Context) error {
	var req MachineCSVApplyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if len(req.Operations) == 0 {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "更新対象が選択されていません。"})
	}

	ctx := c.Request().Context()
	created, updated, err := h.machineUseCase.ApplyCSVOperations(ctx, req.Operations)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"createdCount": created,
		"updatedCount": updated,
		"message":      "CSVインポートが完了しました。",
	})
}

func normalizeMachineColumns(columns []string) []string {
	allow := map[string]bool{
		"name":            true,
		"kana":            true,
		"overview":        true,
		"code":            true,
		"abbreviation":    true,
		"manufacturer_id": true,
		"machine_type":    true,
		"release_date":    true,
		"sort_order":      true,
	}
	seen := map[string]bool{}
	result := make([]string, 0, 9)
	for _, col := range columns {
		normalized := strings.ToLower(strings.TrimSpace(col))
		if allow[normalized] && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	if len(result) == 0 {
		return []string{"name", "kana", "overview", "code", "abbreviation", "manufacturer_id", "machine_type", "release_date", "sort_order"}
	}
	return result
}

func parseMachineCSVRows(csvText string) ([]usecaseAdmin.MachineCSVRowInput, error) {
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
		header := normalizeMachineCSVHeader(raw)
		headers[i] = header
		if header == "id" {
			hasID = true
		}
	}
	if !hasID {
		return nil, fmt.Errorf("id 列が必須です。")
	}

	rows := make([]usecaseAdmin.MachineCSVRowInput, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		row := usecaseAdmin.MachineCSVRowInput{RowNumber: int64(i + 1)}
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
			case "abbreviation":
				v := value
				row.Abbreviation = &v
			case "manufacturer_id":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 64)
				if parseErr != nil {
					continue
				}
				row.ManufacturerID = &num
			case "machine_type":
				v := value
				row.MachineType = &v
			case "release_date":
				v := trimmed
				row.ReleaseDate = &v
			case "sort_order":
				if trimmed == "" {
					continue
				}
				num, parseErr := strconv.ParseInt(trimmed, 10, 32)
				if parseErr != nil {
					continue
				}
				v := int32(num)
				row.SortOrder = &v
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func normalizeMachineCSVHeader(raw string) string {
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
	case "abbreviation", "略称":
		return "abbreviation"
	case "manufacturer_id", "manufacturerid", "メーカーid":
		return "manufacturer_id"
	case "machine_type", "machinetype", "機種種別":
		return "machine_type"
	case "release_date", "releasedate", "発売日":
		return "release_date"
	case "sort_order", "sortorder", "並び順":
		return "sort_order"
	default:
		return ""
	}
}
