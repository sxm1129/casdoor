// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/pressly/goose/v3"
)

// RunGooseMigrations executes pending database migrations using goose.
// The embed.FS must contain a "migrations" directory with SQL migration files.
// In dev mode (runmode=dev), migration errors are logged but not fatal.
// In prod mode, migration failures cause a panic.
func RunGooseMigrations(migrationsFS embed.FS) {
	runMode := conf.GetConfigString("runmode")

	db, err := getGooseDB()
	if err != nil {
		if runMode == "dev" {
			logs.Warning("Goose migration skipped (dev mode): %v", err)
			return
		}
		panic(fmt.Sprintf("Failed to open DB for migrations: %v", err))
	}
	defer db.Close()

	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect(conf.GetConfigString("driverName")); err != nil {
		if runMode == "dev" {
			logs.Warning("Goose dialect error (dev mode): %v", err)
			return
		}
		panic(fmt.Sprintf("Failed to set goose dialect: %v", err))
	}

	if err := goose.Up(db, "migrations"); err != nil {
		if runMode == "dev" {
			logs.Warning("Goose migration error (dev mode): %v", err)
			return
		}
		panic(fmt.Sprintf("Failed to run goose migrations: %v", err))
	}

	logs.Info("Database migrations completed successfully")
}

// getGooseDB opens a raw *sql.DB connection for goose using the same ORM config.
func getGooseDB() (*sql.DB, error) {
	driverName := conf.GetConfigString("driverName")
	dataSourceName := conf.GetConfigDataSourceName()
	dbName := conf.GetConfigString("dbName")

	dsn := dataSourceName
	if driverName == "mysql" {
		dsn = dataSourceName + dbName
	}

	return sql.Open(driverName, dsn)
}
