package contextkeys

type contextKey string

const (
	KeyUserID    contextKey = "user_id"
	KeyRequestID contextKey = "request_id"
)
