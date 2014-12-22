package lib

var (
	ErrUnauthorized = NewJsonError("You need to be signed in to do that!")
	ErrForbidden    = NewJsonError("You do not have permission to do that!")
)

func NewJsonError(msg string) map[string]interface{} {
	return map[string]interface{}{
		"errors": map[string][]string{
			"error": []string{msg},
		},
	}
}
