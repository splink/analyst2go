package db

import (
	_ "github.com/lib/pq"
	"web/src/dbmodel"
	"web/src/util"
)

func Init() {
	InitializeDB(DatabaseConfig{
		Host:       util.Env("POSTGRES_HOST"),
		Name:       util.Env("POSTGRES_DB"),
		Port:       util.Env("POSTGRES_PORT"),
		User:       util.Env("POSTGRES_USER"),
		Password:   util.Env("POSTGRES_PASSWORD"),
		DisableSSL: util.Env("POSTGRES_DISABLE_SSL") == "true",
	})

	DB().MustExec(dbmodel.CreateAppUserTable)
	DB().MustExec(dbmodel.CreateInsightsTable)
	DB().MustExec(dbmodel.CreateDataTable)
	DB().MustExec(dbmodel.CreateAnalysisOptionTable)
	DB().MustExec(dbmodel.CreateAnalysisTable)
	DB().MustExec(dbmodel.CreateCodeTable)
	DB().MustExec(dbmodel.CreateChartTable)
}
