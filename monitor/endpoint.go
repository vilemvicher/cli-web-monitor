package monitor

import (
	"encoding/json"
	"net/http"
)

func CreateGetStatsEndpoint(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		website := r.PathValue("website")
		website = "https://" + website

		query := r.URL.Query()
		page := query.Get("page")

		stats := service.GetStatsForUrl(website, page)

		if stats == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("stats for url not found"))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Accept", "application/json")

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			return
		}
	}
}
