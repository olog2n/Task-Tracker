package tracemiddleware

type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	ProjectIDKey     contextKey = "project_id"
	ProjectMemberKey contextKey = "project_member"
)
