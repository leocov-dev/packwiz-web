package response

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ServerError interface {
	Error() string
	JSON(c *gin.Context)
}

type HttpError struct {
	Code    int
	Message string
}

func New(code int, message string) *HttpError {
	return &HttpError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error) *HttpError {
	return &HttpError{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	}
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *HttpError) JSON(c *gin.Context) {
	c.JSON(e.Code, gin.H{"error": e.Message})
}
