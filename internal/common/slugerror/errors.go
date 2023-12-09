package slugerror

type ErrorType int

const (
	ErrInvalidInput = iota
	ErrUnauthorized
	ErrConflict
	ErrUnprocessable
	ErrInternalServer
)

type SlugError struct {
	error     string
	slug      string
	errorType ErrorType
}

func (s *SlugError) Error() string {
	return s.error
}

func (s *SlugError) Slug() string {
	return s.slug
}

func (s *SlugError) ErrorType() ErrorType {
	return s.errorType
}

func NewSlugError(errorType ErrorType, error string, slug string) *SlugError {
	return &SlugError{
		error:     error,
		slug:      slug,
		errorType: errorType,
	}
}
