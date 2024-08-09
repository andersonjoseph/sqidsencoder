package sqidsencoder

import "fmt"

func typeAssigmentError(fieldName, srcType, dstType string) error {
	return fmt.Errorf("%s with type: %s is not assignable to %s with type: %s.", fieldName, srcType, fieldName, dstType)
}
