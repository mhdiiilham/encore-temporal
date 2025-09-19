package billing

import (
	"encore.dev/storage/sqldb"
)

var billingdb = sqldb.NewDatabase("billings", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
