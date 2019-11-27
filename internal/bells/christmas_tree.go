package bells

import "net/http"

// ChristmasTree adds a Christmas Tree to a header!
func ChristmasTree(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("X-Christmas-Tree", "🎄")
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
