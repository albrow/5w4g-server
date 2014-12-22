package lib

var (
	ErrUnauthorized = map[string]interface{}{
		"errors": map[string][]string{
			"error": []string{"You need to be signed in to do that!"},
		},
	}
	ErrForbidden = map[string]interface{}{
		"errors": map[string][]string{
			"error": []string{"You do not have permission to do that!"},
		},
	}
)
