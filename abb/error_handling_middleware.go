package abb

import (
	"fmt"

	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"
)

type ErrorHandlingMiddleware struct {
}

func NewErrorHandlingMiddleware() *ErrorHandlingMiddleware {
	return &ErrorHandlingMiddleware{}
}

func (m *ErrorHandlingMiddleware) Invoke(c *napnap.Context, next napnap.HandlerFunc) {
	defer func() {
		// we only handle error for bifrost application and don't handle can't error from upstream.
		if r := recover(); r != nil {
			// bad request.  http status code is 400 series.
			appError, ok := r.(app.AppError)
			if ok {
				if appError.ErrorCode == "not_found" {
					c.JSON(404, appError)
					return
				}
				c.JSON(400, appError)
				return
			}

			// unknown error.  http status code is 500 series.
			logger := log.StackTrace()
			customFields := log.Fields{
				"url": c.Request.RequestURI,
			}
			err, ok := r.(error)
			if !ok {
				if err == nil {
					err = fmt.Errorf("%v", r)
				} else {
					err = fmt.Errorf("%v", err)
				}
			}
			logger.WithFields(customFields).Errorf("unknown error: %v", err)

			appError = app.AppError{
				ErrorCode: "unknown_error",
				Message:   err.Error(),
			}
			c.JSON(500, appError)
		}
	}()
	next(c)
}
