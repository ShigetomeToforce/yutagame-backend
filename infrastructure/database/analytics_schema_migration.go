package database

import (
	"fmt"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

func DropLegacyAnalyticsColumns(db *gorm.DB) error {
	if err := dropColumnsIfExists(db, &model.SearchLog{}, []string{
		"event_source",
		"path",
		"method",
		"status_code",
		"user_agent",
		"referrer",
	}); err != nil {
		return fmt.Errorf("drop legacy search_logs columns: %w", err)
	}

	if err := dropColumnsIfExists(db, &model.GameViewLog{}, []string{
		"event_source",
		"path",
		"method",
		"status_code",
		"user_agent",
		"referrer",
	}); err != nil {
		return fmt.Errorf("drop legacy game_view_logs columns: %w", err)
	}

	return nil
}

func dropColumnsIfExists(db *gorm.DB, modelRef any, columns []string) error {
	migrator := db.Migrator()
	if !migrator.HasTable(modelRef) {
		return nil
	}

	for _, column := range columns {
		if !migrator.HasColumn(modelRef, column) {
			continue
		}
		if err := migrator.DropColumn(modelRef, column); err != nil {
			return fmt.Errorf("drop column %s: %w", column, err)
		}
	}

	return nil
}
