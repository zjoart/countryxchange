package countries

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
func RegisterRoutes(r *mux.Router, db *sql.DB, isProduction bool) {
	r.HandleFunc("/countries/refresh", func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// handler-level structured log: calling refresh service
		logger.Info("handler: calling Refresh service", logger.Fields{
			"action":      "countries.refresh",
			"remote_addr": req.RemoteAddr,
			"user_agent":  req.UserAgent(),
			"db_present":  db != nil,
		})

		res, err := Refresh(ctx, db)
		if err != nil {
			// validation error
			if verr, ok := err.(*ValidationError); ok {
				logger.Warn("handler: validation failed", logger.Fields{"errors": verr.Errors})
				writeError(w, http.StatusBadRequest, "Validation failed", verr.Errors)
				return
			}
			// external API error
			if _, ok := err.(ExternalError); ok {
				logger.Warn("handler: refresh failed - external API", logger.Fields{"error": err.Error()})
				writeError(w, http.StatusServiceUnavailable, "External data source unavailable", err.Error())
				return
			}
			logger.Error("handler: refresh failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}

		logger.Info("handler: refresh completed", logger.Fields{"total_processed": res.Total, "last_refreshed_at": res.LastRefreshed.Format(time.RFC3339)})
		writeJSON(w, http.StatusOK, map[string]interface{}{"message": "refreshed", "total": res.Total, "last_refreshed_at": res.LastRefreshed.Format(time.RFC3339)})
	}).Methods("POST")

	r.HandleFunc("/countries", func(w http.ResponseWriter, req *http.Request) {
		// sanitize query keys to defensively handle malformed clients that send
		// keys like "?currency" (extra '?'). Trim any leading '?' from keys.
		raw := req.URL.Query()
		q := make(map[string][]string)
		for k, v := range raw {
			nk := strings.TrimLeft(k, "?")
			q[nk] = v
		}
		// helper to mimic url.Values.Get
		get := func(key string) string {
			if vals, ok := q[key]; ok && len(vals) > 0 {
				return vals[0]
			}
			return ""
		}

		region := get("region")
		currency := get("currency")
		sort := get("sort")
		logger.Info("handler: listing countries", logger.Fields{"region": region, "currency": currency, "sort": sort})
		list, err := GetAll(db, region, currency, sort)
		if err != nil {
			logger.Error("get all countries failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		logger.Info("handler: listed countries", logger.Fields{"count": len(list)})
		writeJSON(w, http.StatusOK, list)
	}).Methods("GET")

	r.HandleFunc("/countries/image", func(w http.ResponseWriter, req *http.Request) {
		path := filepath.Join("cache", "summary.png")
		logger.Info("handler: serve summary image", logger.Fields{"path": path})
		if _, err := os.Stat(path); err != nil {
			logger.Warn("handler: summary image not found", logger.Fields{"path": path})
			writeError(w, http.StatusNotFound, "Summary image not found", nil)
			return
		}
		http.ServeFile(w, req, path)
	}).Methods("GET")

	r.HandleFunc("/countries/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		logger.Info("handler: get country by name", logger.Fields{"name": name, "remote_addr": req.RemoteAddr})
		c, err := GetByName(db, name)
		if err != nil {
			if err == ErrNotFound {
				logger.Debug("handler: country not found", logger.Fields{"name": name})
				writeError(w, http.StatusNotFound, "Country not found", nil)
				return
			}
			logger.Error("handler: get country failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		logger.Info("handler: get country success", logger.Fields{"name": c.Name, "id": c.ID})
		writeJSON(w, http.StatusOK, c)
	}).Methods("GET")

	r.HandleFunc("/countries/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		logger.Info("handler: delete country by name", logger.Fields{"name": name, "remote_addr": req.RemoteAddr})
		ok, err := DeleteByName(db, name)
		if err != nil {
			logger.Error("handler: delete country failed", logger.WithError(err))
			writeError(w, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		if !ok {
			logger.Debug("handler: delete country not found", logger.Fields{"name": name})
			writeError(w, http.StatusNotFound, "Country not found", nil)
			return
		}
		logger.Info("handler: delete country success", logger.Fields{"name": name})
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
	}).Methods("DELETE")

	r.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		logger.Info("handler: status check")
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
		logger.Info("handler: status response", logger.Fields{"total_countries": total, "last_refreshed_at": lastStr})
		writeJSON(w, http.StatusOK, map[string]interface{}{"total_countries": total, "last_refreshed_at": lastStr})
	}).Methods("GET")

	if !isProduction {
		// Drop tables endpoint - BE CAREFUL WITH THIS IN PRODUCTION!
		r.HandleFunc("/drop-tables", func(w http.ResponseWriter, req *http.Request) {
			logger.Warn("handler: dropping all tables", logger.Fields{"remote_addr": req.RemoteAddr})

			if err := DropTables(db); err != nil {
				logger.Error("handler: drop tables failed", logger.WithError(err))
				writeError(w, http.StatusInternalServerError, "Failed to drop tables", nil)
				return
			}

			logger.Info("handler: tables dropped successfully")
			writeJSON(w, http.StatusOK, map[string]string{"message": "Tables dropped successfully"})
		}).Methods("POST")
	}
}
