package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxBody = 1 << 20 // 1 MiB

func ReadJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var syn *json.SyntaxError
		switch {
		case errors.As(err, &syn):
			return fmt.Errorf("malformed JSON at byte %d", syn.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("empty body")
		default:
			return err
		}
	}
	if dec.More() {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
