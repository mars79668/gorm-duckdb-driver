package duckdb

import (
	"log"

	"gorm.io/gorm"
)

// CustomRowQuery is a debugging version of GORM's RowQuery callback
func CustomRowQuery(db *gorm.DB) {
	log.Printf(" CustomRowQuery called")
	log.Printf(" db.Error: %v", db.Error)
	log.Printf(" db.DryRun: %t", db.DryRun)

	if db.Error == nil {
		log.Printf(" No error, calling BuildQuerySQL")
		// This is what GORM's BuildQuerySQL does for Raw queries
		if db.Statement.SQL.Len() == 0 {
			log.Printf(" SQL is empty, this shouldn't happen for Raw() queries")
		}

		// Check for DryRun or Error before proceeding
		if db.DryRun || db.Error != nil {
			log.Printf(" DryRun=%t or Error=%v, returning early", db.DryRun, db.Error)
			return
		}

		log.Printf(" Checking for 'rows' setting")
		if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
			log.Printf(" isRows=true, calling QueryContext")
			db.Statement.Settings.Delete("rows")
			db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
		} else {
			log.Printf(" isRows=false or not found, calling QueryRowContext")
			log.Printf(" SQL: %s", db.Statement.SQL.String())
			log.Printf(" Vars: %v", db.Statement.Vars)
			log.Printf(" ConnPool type: %T", db.Statement.ConnPool)

			result := db.Statement.ConnPool.QueryRowContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			log.Printf(" QueryRowContext returned: %v (nil: %t)", result, result == nil)

			db.Statement.Dest = result
			log.Printf(" After assignment - Statement.Dest: %v (nil: %t)", db.Statement.Dest, db.Statement.Dest == nil)
		}

		log.Printf(" Setting RowsAffected to -1")
		db.RowsAffected = -1
	} else {
		log.Printf(" db.Error is not nil: %v", db.Error)
	}
}
