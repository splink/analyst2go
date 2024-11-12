package service

import (
	"fmt"
	"github.com/lib/pq"
	"log"
	"strings"
	"time"
	"web/src/db"
	"web/src/model"
	"web/src/util"
)

func SaveInsightData(insightID int64, dataFile model.DataFile) error {
	// Generate S3 key (unique identifier) based on insightID and current timestamp
	s3key := fmt.Sprintf("%d%d%s", insightID, time.Now().Unix(), dataFile.Ext)

	filePath, err := util.UploadToS3(s3key, dataFile.Data)
	if err != nil {
		log.Println("Failed to upload file to S3:", err)
		return err
	}

	// Convert Headers and FirstRows to database-friendly formats
	headers := strings.Join(dataFile.Headers, ",")
	firstRows := make([]string, len(dataFile.FirstRows))
	for i, row := range dataFile.FirstRows {
		firstRows[i] = strings.Join(row, ",")
	}

	query := `
		INSERT INTO insight_data (insight_id, s3key, file_size, file_extension, uploaded_at, headers, first_rows)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (insight_id) DO UPDATE SET
			s3key = EXCLUDED.s3key,
			file_size = EXCLUDED.file_size,
			file_extension = EXCLUDED.file_extension,
			uploaded_at = EXCLUDED.uploaded_at,
			headers = EXCLUDED.headers,
			first_rows = EXCLUDED.first_rows;
	`

	_, err = db.DB().Exec(query,
		insightID,
		s3key,
		len(dataFile.Data),
		dataFile.Ext,
		time.Now(),
		headers,
		pq.Array(firstRows))
	if err != nil {
		return fmt.Errorf("failed to insert or update insight_data: %w", err)
	}

	log.Printf("Data saved to S3 at %s and database updated for insight_id %d", filePath, insightID)
	return nil
}
