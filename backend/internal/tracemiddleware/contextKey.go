package tracemiddleware

type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	UserEmailKey     contextKey = "user_email"
	UserNameKey      contextKey = "user_name"
	ProjectIDKey     contextKey = "project_id"
	ProjectMemberKey contextKey = "project_member"
)
