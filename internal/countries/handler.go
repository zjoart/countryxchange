package countries

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/zjoart/countryxchange/pkg/logger"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, details interface{}) {
	payload := map[string]interface{}{"error": msg}
	if details != nil {
		payload["details"] = details
	}
	writeJSON(w, status, payload)
}

// RegisterRoutes mounts country endpoints onto router
func RegisterRoutes(r *mux.Router, db *sql.DB) {
	r.HandleFunc("/countries/refresh", func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		res, err := Refresh(ctx, db)
		if err != nil {
			// external API error
			if _, ok := err.(ExternalError); ok {
				writeError(w, http.StatusServiceUnavailable, "External data source unavailable", map[string]string{"details": err.Error()})
				return
			}
			logger.Error("refresh failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"message": "refreshed", "total": res.Total, "last_refreshed_at": res.LastRefreshed.Format(time.RFC3339)})
	}).Methods("POST")

	r.HandleFunc("/countries", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		region := q.Get("region")
		currency := q.Get("currency")
		sort := q.Get("sort")
		list, err := GetAll(db, region, currency, sort)
		if err != nil {
			logger.Error("get all countries failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		writeJSON(w, http.StatusOK, list)
	}).Methods("GET")

	r.HandleFunc("/countries/image", func(w http.ResponseWriter, req *http.Request) {
		path := filepath.Join("cache", "summary.png")
		if _, err := os.Stat(path); err != nil {
			writeError(w, http.StatusNotFound, "Summary image not found", nil)
			return
		}
		http.ServeFile(w, req, path)
	}).Methods("GET")

	r.HandleFunc("/countries/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		c, err := GetByName(db, name)
		if err != nil {
			if err == ErrNotFound {
				writeError(w, http.StatusNotFound, "Country not found", nil)
				return
			}
			logger.Error("get country failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		writeJSON(w, http.StatusOK, c)
	}).Methods("GET")

	r.HandleFunc("/countries/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		ok, err := DeleteByName(db, name)
		if err != nil {
			logger.Error("delete country failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "Country not found", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
	}).Methods("DELETE")

	r.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		total, err := TotalCount(db)
		if err != nil {
			logger.Error("status failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		last, err := GetLastRefreshed(db)
		if err != nil {
			logger.Error("status failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		var lastStr *string
		if last != nil {
			s := last.UTC().Format(time.RFC3339)
			lastStr = &s
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"total_countries": total, "last_refreshed_at": lastStr})
	}).Methods("GET")
}
