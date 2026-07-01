package httpx

import (
	"net/http"
	"strconv"
	"time"
)

// QueryInt returns an integer query parameter with a default value.
func QueryInt(r *http.Request, key string, defaultVal int) int {
	if val := r.URL.Query().Get(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// ParseTime parses an ISO 8601 timestamp string.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
