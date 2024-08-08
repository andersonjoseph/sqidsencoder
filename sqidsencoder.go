package sqidsencoder

import (
	"fmt"
	"reflect"
)

type sqidsInterface interface {
	Encode(numbers []uint64) (string, error)
	Decode(id string) []uint64
}

type sqidsencoder struct {
	sqids sqidsInterface
}

func New(s sqidsInterface) sqidsencoder {
	return sqidsencoder{
		sqids: s,
	}
}

func (enc sqidsencoder) Encode(src any, dst any) error {
	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)

	destVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcType.NumField(); i++ {
		currentDstField := destVal.FieldByName(srcType.Field(i).Name)
		currentSrcField := srcVal.FieldByName(srcType.Field(i).Name)

		if op, ok := srcType.Field(i).Tag.Lookup("sqids"); ok && op == "encode" {
			encodedID, err := enc.sqids.Encode([]uint64{uint64(srcVal.Field(i).Int())})

			if err != nil {
				return err
			}

			currentDstField.SetString(encodedID)
			continue
		}

		if !currentSrcField.Type().AssignableTo(currentDstField.Type()) {
			fieldName := srcType.Field(i).Name
			srcTypeName := currentSrcField.Type().Name()
			dstTypeName := currentDstField.Type().Name()
			return fmt.Errorf("%s with type: %s is not assignable to %s with type: %s.", fieldName, srcTypeName, fieldName, dstTypeName)
		}

		currentDstField.Set(currentSrcField)
	}

	return nil
}

func (enc sqidsencoder) Decode(src any, dst any) error {
	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)

	destVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcType.NumField(); i++ {
		fieldName := srcType.Field(i).Name

		if op, ok := srcType.Field(i).Tag.Lookup("sqids"); ok && op == "decode" {
			decodedID := enc.sqids.Decode(srcVal.Field(i).String())[0]

			destVal.FieldByName(fieldName).SetInt(int64(decodedID))
			continue
		}

		currentDstField := destVal.FieldByName(srcType.Field(i).Name)
		currentSrcField := srcVal.FieldByName(srcType.Field(i).Name)

		if !currentSrcField.Type().AssignableTo(currentDstField.Type()) {
			fieldName := srcType.Field(i).Name
			srcTypeName := currentSrcField.Type().Name()
			dstTypeName := currentDstField.Type().Name()
			return fmt.Errorf("%s with type: %s is not assignable to %s with type: %s.", fieldName, srcTypeName, fieldName, dstTypeName)
		}

		destVal.FieldByName(fieldName).Set(srcVal.FieldByName(fieldName))
	}

	return nil
}
