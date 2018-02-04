package retryhttp

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

//go:generate counterfeiter . Retryer

type Retryer interface {
	IsRetryable(err error) bool
}

type DefaultRetryer struct{}

func (r *DefaultRetryer) IsRetryable(err error) bool {
	if neterr, ok := err.(net.Error); ok {
		if neterr.Temporary() {
			return true
		}
	}

	s := err.Error()
	for _, retryableError := range defaultRetryableErrors {
		if strings.HasSuffix(s, retryableError.Error()) {
			return true
		}
	}

	return false
}

var defaultRetryableErrors = []error{
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
	syscall.ETIMEDOUT,
	errors.New("i/o timeout"),
	errors.New("no such host"),
	errors.New("remote error: handshake failure"),
}
