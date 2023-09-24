package adapter

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

var authCodeTest = GetEnv("AUTH_TOKEN", "test")

func (p *MongoDBAdapter) handleAuthRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) bool {
	path := r.URL.Path
	if strings.HasPrefix(path, "/_health") {
		w.WriteHeader(200)
		return false
	}
	apiKey := strings.Replace(r.Header.Get("authorization"), "Bearer ", "", 1)
	if apiKey != authCodeTest {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
