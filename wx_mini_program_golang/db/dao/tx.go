package dao

import (
	"wxcloudrun-golang/db"

	"gorm.io/gorm"
)

// Transaction wraps a function in a database transaction.
// It uses the global db.Get() instance to start a transaction.
// The provided function 'fc' receives a *gorm.DB object 'tx',
// which should be used for all database operations within the transaction.
func Transaction(fc func(tx *gorm.DB) error) error {
	return db.Get().Transaction(fc)
}
