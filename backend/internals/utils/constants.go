package utils

// constants.go - Centralized constants for the backend application.

const (
	// JWTSecretKeyEnv is the environment variable name for the JWT secret
	JWTSecretKeyEnv = "JWT_SECRET"

	// DefaultJWTSecret is the fallback secret if env is not set (should be overridden in production)
	DefaultJWTSecret = "itiswhatitis"

	// UsersFileName is the name of the users CSV file
	UsersFileName = "users.csv"

	// UsersDataDir is the directory where user data is stored
	UsersDataDir = "data"

	// DefaultAdminUsername is the default admin username
	DefaultAdminUsername = "admin"

	// DefaultAdminPassword is the default admin password
	DefaultAdminPassword = "password"

	// DefaultPort is the default port for the server
	DefaultPort = "8080"

	// DefaultMessageLimit is the default number of messages to fetch
	DefaultMessageLimit = 5

	// DefaultMessageSort is the default sort order for messages
	DefaultMessageSort = "newest"

	// StatusSuccess is the status for successful operations
	StatusSuccess = "success"

	// StatusError is the status for error operations
	StatusError = "error"
)
