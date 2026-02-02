package domain

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Meta struct {
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	Total      int64  `json:"total,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	RequestID  string `json:"request_id"`
}

type SearchData struct {
	Items []Content `json:"items"`
}

type SearchRequest struct {
	Query        string   `json:"query"`
	Tags         []string `json:"tags"`
	ContentTypes []string `json:"types"`
	OrderBy      string   `json:"orderBy"`
	Page         int      `json:"page"`
	PerPage      int      `json:"perPage"`
}

func (r *SearchRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PerPage <= 0 {
		r.PerPage = 20
	}
	if r.PerPage > 100 {
		r.PerPage = 100
	}
	if r.OrderBy == "" {
		r.OrderBy = "relevant_score"
	}
}

func NewSuccessResponse(data interface{}, meta *Meta) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

func NewErrorResponse(code, message string, requestID string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Meta: &Meta{
			RequestID: requestID,
		},
	}
}
