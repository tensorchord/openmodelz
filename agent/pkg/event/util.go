package event

import (
	"database/sql"
)

func NullStringBuilder(String string, Valid bool) sql.NullString {
	return sql.NullString{String: String, Valid: Valid}
}
