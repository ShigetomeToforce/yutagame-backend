package admin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var ErrCodeAlreadyExists = errors.New("code is already in use")

func validateDuplicateCode[T any](
	ctx context.Context,
	code string,
	excludeID int64,
	finder func(context.Context, string) (*T, error),
) error {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil
	}

	existing, err := finder(ctx, trimmedCode)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	value := reflect.ValueOf(existing)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}

	idField := value.FieldByName("ID")
	if !idField.IsValid() || !idField.CanInt() {
		return nil
	}

	if idField.Int() != excludeID {
		return fmt.Errorf("%w: %s", ErrCodeAlreadyExists, trimmedCode)
	}

	return nil
}
