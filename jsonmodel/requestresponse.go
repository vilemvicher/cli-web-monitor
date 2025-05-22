package jsonmodel

import "time"

type RequestResponse struct {
	Date       time.Time `json:"date"`
	OK         bool      `json:"ok"`
	DurationMs float64   `json:"duration"`
}
