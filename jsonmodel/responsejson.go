package jsonmodel

type WebsiteResponse struct {
	Pagination Pagination        `json:"pagination"`
	Requests   []RequestResponse `json:"requests"`
}
type Pagination struct {
	Page       int `json:"page"`
	TotalPages int `json:"totalPages"`
	Items      int `json:"items"`
	TotalItems int `json:"totalItems"`
}
