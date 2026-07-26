package handlers

// Auth handles authentication-related HTTP requests.
type Auth struct {
	auth  AuthService
	users UserService
}

// NewAuth creates an authentication handler with explicit service dependencies.
func NewAuth(authService AuthService, userService UserService) *Auth {
	return &Auth{auth: authService, users: userService}
}
