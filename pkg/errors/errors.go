package errors

type MultipleFormatError struct {
	Message string
}

type ImageError struct {
	Message string
}

type GrafanaAPIVersionSupportError struct {
	Message string
}

func (e *MultipleFormatError) Error() string {
	return e.Message
}

func (e *ImageError) Error() string {
	return e.Message
}

func (e *GrafanaAPIVersionSupportError) Error() string {
	return e.Message
}
