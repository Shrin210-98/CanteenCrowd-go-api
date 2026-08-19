// internal/utils/db_errors.go
package utils

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// HandleDBError handles database errors and sends appropriate HTTP responses
// If notFoundMsg is provided and the error is pgx.ErrNoRows, it will use that message
func HandleDBError(w http.ResponseWriter, err error, operation string, notFoundMsg ...string) {
	if err == nil {
		return
	}

	// Handle pgx.ErrNoRows first (most common - resource not found)
	if errors.Is(err, pgx.ErrNoRows) {
		msg := "Resource not found"
		if len(notFoundMsg) > 0 && notFoundMsg[0] != "" {
			msg = notFoundMsg[0]
		}
		log.Printf("%s: %s", operation, msg)
		ErrorResponse(w, http.StatusNotFound, msg)
		return
	}

	// Handle PostgreSQL specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			log.Printf("%s: Unique violation on constraint %s - %s", operation, pgErr.ConstraintName, pgErr.Detail)
			ErrorResponse(w, http.StatusConflict, "A record with this information already exists")
			return

		case "23503": // foreign_key_violation
			log.Printf("%s: Foreign key violation on constraint %s - %s", operation, pgErr.ConstraintName, pgErr.Detail)
			ErrorResponse(w, http.StatusConflict, "Operation cannot be completed: related resource does not exist or is in use")
			return

		case "23502": // not_null_violation
			log.Printf("%s: Not null violation on column %s - %s", operation, pgErr.ColumnName, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "A required field is missing")
			return

		case "23514": // check_violation
			log.Printf("%s: Check constraint violation - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Invalid data provided")
			return

		case "22P02": // invalid_text_representation (often UUID parsing)
			log.Printf("%s: Invalid input format - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Invalid input format")
			return

		case "22001": // string_data_right_truncation
			log.Printf("%s: String data too long - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Input data is too long")
			return

		case "22003": // numeric_value_out_of_range
			log.Printf("%s: Numeric value out of range - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Numeric value is out of range")
			return

		case "22007": // invalid_datetime_format
			log.Printf("%s: Invalid datetime format - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Invalid date/time format")
			return

		case "22008": // datetime_field_overflow
			log.Printf("%s: Datetime field overflow - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Date/time value out of range")
			return

		case "22P05": // untranslatable_character
			log.Printf("%s: Untranslatable character - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusBadRequest, "Input contains invalid characters")
			return

		case "23506": // exclusion_violation
			log.Printf("%s: Exclusion constraint violation - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusConflict, "Operation violates data constraints")
			return

		case "40001": // serialization_failure
			log.Printf("%s: Serialization failure - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusConflict, "Concurrent modification detected, please retry")
			return

		case "40P01": // deadlock_detected
			log.Printf("%s: Deadlock detected - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusConflict, "Operation conflict detected, please retry")
			return

		case "57014": // query_canceled
			log.Printf("%s: Query canceled - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusRequestTimeout, "Request timed out")
			return

		case "53P02": // too_many_connections
			log.Printf("%s: Too many connections - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
			return

		case "57P01": // admin_shutdown
			log.Printf("%s: Database shutting down - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
			return

		case "57P03": // cannot_connect_now
			log.Printf("%s: Cannot connect now - %s", operation, pgErr.Detail)
			ErrorResponse(w, http.StatusServiceUnavailable, "Database is starting up, please try again")
			return

		default:
			log.Printf("%s: Database error code %s - %s", operation, pgErr.Code, pgErr.Message)
			ErrorResponse(w, http.StatusInternalServerError, "Database operation failed")
			return
		}
	}

	// Handle connection timeout errors
	if pgconn.Timeout(err) {
		log.Printf("%s: Database connection timeout", operation)
		ErrorResponse(w, http.StatusRequestTimeout, "Database request timed out")
		return
	}

	// Handle connection errors
	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		log.Printf("%s: Database connection error: %v", operation, connErr)
		ErrorResponse(w, http.StatusServiceUnavailable, "Database service is unavailable")
		return
	}

	// Default error handling
	log.Printf("%s: Unexpected database error: %v", operation, err)
	ErrorResponse(w, http.StatusInternalServerError, "An internal error occurred")
}

// HandleNotFound is a helper for handling not found errors specifically
// Returns true if the error was a "not found" error and was handled
func HandleNotFound(w http.ResponseWriter, err error, message string) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		ErrorResponse(w, http.StatusNotFound, message)
		return true
	}
	return false
}

// Helper functions to check specific error types
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

func IsNotNullViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23502"
	}
	return false
}

func IsCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23514"
	}
	return false
}
