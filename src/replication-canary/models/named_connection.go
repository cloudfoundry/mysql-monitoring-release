package models

import "database/sql"

type NamedConnection struct {
	Name       string
	Connection *sql.DB
}
