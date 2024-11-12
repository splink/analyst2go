package service

import (
	"fmt"
	"time"
	"web/src/db"
)

func CreateInsight(userID int64) (int64, error) {
	createdAt := time.Now()
	updatedAt := createdAt
	isDeleted := false

	query := `
		INSERT INTO insights (user_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4)
		RETURNING insight_id;
	`

	var insightID int64
	err := db.DB().QueryRow(query, userID, createdAt, updatedAt, isDeleted).Scan(&insightID)
	if err != nil {
		return 0, fmt.Errorf("failed to create insight: %w", err)
	}

	return insightID, nil
}
