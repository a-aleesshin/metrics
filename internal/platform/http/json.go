package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// IsJSON сообщает, имеет ли запрос Content-Type application/json.
func IsJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

	if err != nil || mediaType != "application/json" {
		return false
	}

	return true
}

// DecodeJSON декодирует тело запроса в dst; неизвестные поля
// и несколько JSON-объектов подряд считаются ошибкой.
func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("invalid json: multiple objects")
	}

	return nil
}

// WriteJSON пишет payload в ответ как JSON с указанным HTTP-статусом.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
