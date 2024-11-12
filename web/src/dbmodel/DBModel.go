package dbmodel

import "time"

type AppUser struct {
	UserID    int64     `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Insight struct {
	InsightID int64     `json:"insight_id" db:"insight_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
}

type InsightData struct {
	InsightID     int64     `json:"insight_id" db:"insight_id"`
	S3key         string    `json:"s3key" db:"s3key"`
	FileSize      int       `json:"file_size" db:"file_size"`
	FileExtension string    `json:"file_extension" db:"file_extension"`
	UploadedAt    time.Time `json:"uploaded_at" db:"uploaded_at"`
	Headers       string    `json:"headers" db:"headers"`
	FirstRows     []string  `json:"first_rows" db:"first_rows"`
}

type AnalysisOption struct {
	OptionID    int64     `json:"option_id" db:"option_id"`
	InsightID   int64     `json:"insight_id" db:"insight_id"`
	Name        string    `json:"name" db:"name"`
	ChartType   string    `json:"chart_type" db:"chart_type"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type InsightAnalysis struct {
	InsightID        int64     `json:"insight_id" db:"insight_id"`
	SelectedOptionID int64     `json:"selected_option_id" db:"selected_option_id"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type InsightCode struct {
	InsightID int64     `json:"insight_id" db:"insight_id"`
	Code      string    `json:"code" db:"code"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type InsightChart struct {
	InsightID int64     `json:"insight_id" db:"insight_id"`
	ChartData string    `json:"chart_data" db:"chart_data"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

var CreateAppUserTable = `
CREATE TABLE IF NOT EXISTS app_user (
    user_id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_app_user_email ON app_user (email);`

var CreateInsightsTable = `
CREATE TABLE IF NOT EXISTS insights (
    insight_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES app_user(user_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_insights_user_id ON insights (user_id);`

var CreateDataTable = `
CREATE TABLE IF NOT EXISTS insight_data (
    insight_id BIGINT PRIMARY KEY REFERENCES insights(insight_id),
    s3key TEXT NOT NULL,
    file_size INT,
    file_extension TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    headers TEXT,
    first_rows TEXT[]
);

DROP TRIGGER IF EXISTS trg_data_update ON insight_data;
DROP FUNCTION IF EXISTS on_data_update;

CREATE OR REPLACE FUNCTION on_data_update() RETURNS TRIGGER AS $$
BEGIN
    -- Only delete dependent records if headers or first_rows are updated
    IF NEW.headers IS DISTINCT FROM OLD.headers OR NEW.first_rows IS DISTINCT FROM OLD.first_rows THEN
        DELETE FROM insight_analysis WHERE insight_id = NEW.insight_id;
        DELETE FROM insight_code WHERE insight_id = NEW.insight_id;
        DELETE FROM insight_chart WHERE insight_id = NEW.insight_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_data_update
BEFORE UPDATE ON insight_data
FOR EACH ROW EXECUTE FUNCTION on_data_update();`

var CreateAnalysisOptionTable = `
CREATE TABLE IF NOT EXISTS analysis_options (
    option_id BIGSERIAL PRIMARY KEY,
    insight_id BIGINT REFERENCES insights(insight_id),
    name TEXT NOT NULL,
    chart_type TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_analysis_options_insight_id ON analysis_options (insight_id);`

var CreateAnalysisTable = `
CREATE TABLE IF NOT EXISTS insight_analysis (
    insight_id BIGINT PRIMARY KEY REFERENCES insights(insight_id),
    selected_option_id BIGINT REFERENCES analysis_options(option_id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS trg_analysis_update ON insight_analysis;
DROP FUNCTION IF EXISTS on_analysis_update;

CREATE OR REPLACE FUNCTION on_analysis_update() RETURNS TRIGGER AS $$
BEGIN
    -- Only delete dependent records if selected_option_id changes
    IF NEW.selected_option_id IS DISTINCT FROM OLD.selected_option_id THEN
        DELETE FROM insight_code WHERE insight_id = NEW.insight_id;
        DELETE FROM insight_chart WHERE insight_id = NEW.insight_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_analysis_update
BEFORE UPDATE ON insight_analysis
FOR EACH ROW EXECUTE FUNCTION on_analysis_update();`

var CreateCodeTable = `
CREATE TABLE IF NOT EXISTS insight_code (
    insight_id BIGINT PRIMARY KEY REFERENCES insights(insight_id),
    code TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS trg_code_update ON insight_code;
DROP FUNCTION IF EXISTS on_code_update;

CREATE OR REPLACE FUNCTION on_code_update() RETURNS TRIGGER AS $$
BEGIN
    -- Only delete dependent records if code changes
    IF NEW.code IS DISTINCT FROM OLD.code THEN
        DELETE FROM insight_chart WHERE insight_id = NEW.insight_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_code_update
BEFORE UPDATE ON insight_code
FOR EACH ROW EXECUTE FUNCTION on_code_update();`

var CreateChartTable = `
CREATE TABLE IF NOT EXISTS insight_chart (
    insight_id BIGINT PRIMARY KEY REFERENCES insights(insight_id),
    chart_data TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`
