package errors

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

func IgnoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}
