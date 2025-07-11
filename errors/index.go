package utilsErrors

import (
	"errors"
	"runtime"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func NewServiceError(code, userMsg string, httpStatus int) *ServiceError {
	return &ServiceError{
		code:     code,
		message:  userMsg,
		httpCode: httpStatus,
		err:      errors.New(code),
	}
}

type ErrorResponseHttp struct {
	Error      string            `json:"error"`
	Message    string            `json:"message"`
	ErrorsForm map[string]string `json:"errorsForm"`
}

type ServiceError struct {
	code       string
	err        error
	errorType  string
	message    string
	httpCode   int
	ErrorsForm map[string]string `json:"errorsForm"`
	Stack      []byte            // Optional stack trace
}

func (e *ServiceError) Code() string {
	return e.code
}

func (e *ServiceError) ErrorCode() error {
	return errors.New(e.code)
}

func (e *ServiceError) Error() error {
	return e.err
}

func (e *ServiceError) ErrorType() string {
	return e.errorType
}

func (e *ServiceError) Message() string {
	return e.message
}

func (e *ServiceError) HttpCode() int {
	return e.httpCode
}

func (e *ServiceError) SetError(err error) {
	e.err = err
}

func (e *ServiceError) GetHttpResponseError() ErrorResponseHttp {
	return ErrorResponseHttp{
		Error:      e.code,
		Message:    e.message,
		ErrorsForm: e.ErrorsForm,
	}
}

// Helper to wrap existing errors
func WrapError(baseErr *ServiceError, cause error) *ServiceError {
	return &ServiceError{
		code:     baseErr.Code(),
		message:  baseErr.Message(),
		httpCode: baseErr.httpCode,
		err:      cause,
		//Stack:    baseErr.Stack, // Or capture new stack here
	}
}

// WithStack captures stack trace at call time
func (e *ServiceError) WithStack() *ServiceError {
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	e.Stack = stack[:n]
	return e
}

// Add context to errors
func (e *ServiceError) WithCause(cause error) *ServiceError {
	return &ServiceError{
		code:     e.code,
		message:  e.message,
		httpCode: e.httpCode,
		err:      cause,
		Stack:    e.Stack, // Preserve existing stack if any
	}
}

// Add context to errors
func (e *ServiceError) WithBadInputItems(cause error) *ServiceError {
	errorsForm := make(map[string]string)
	if e, ok := cause.(validation.Errors); ok {
		for field, fieldErr := range e {
			errorsForm[field] = fieldErr.Error()
		}
	}
	return &ServiceError{
		code:       e.code,
		message:    e.message,
		httpCode:   e.httpCode,
		err:        cause,
		Stack:      e.Stack, // Preserve existing stack if any
		ErrorsForm: errorsForm,
	}
}
