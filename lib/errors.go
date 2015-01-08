package lib

var (
	ErrUnauthorized = NewJsonError("You need to be signed in to do that!")
	ErrForbidden    = NewJsonError("You do not have permission to do that!")
)

func NewJsonError(msg string) map[string]string {
	return map[string]string{
		"error": msg,
	}
}
