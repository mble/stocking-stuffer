package bells

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UserID adds a SHA256 hash of the username to the logs when passed as a body entry.
func UserID(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var fields interface{}
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		err := json.Unmarshal(body, &fields)
		if err == nil {
			username := fields.(map[string]interface{})["username"]
			if username != nil {
				hashedUsername := fmt.Sprintf("%x", sha256sum(username))
				LogEntrySetField(r, "user_id", hashedUsername)
			}
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func sha256sum(data interface{}) [32]byte {
	bytes := []byte(data.(string))
	sum := sha256.Sum256(bytes)
	return sum
}
