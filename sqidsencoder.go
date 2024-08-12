package sqidsencoder

import (
	"fmt"
	"reflect"
)

type encoderOperation string

const (
	SQIDS_TAG                  = "sqids"
	ENCODE    encoderOperation = "encode"
	DECODE    encoderOperation = "decode"
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
	return enc.buildDstStruct(src, dst, ENCODE)
}

func (enc sqidsencoder) Decode(src any, dst any) error {
	return enc.buildDstStruct(src, dst, DECODE)
}

func (enc sqidsencoder) buildDstStruct(src any, dst any, op encoderOperation) error {
	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)

	if srcVal.Kind() != reflect.Struct {
		return fmt.Errorf("src must be a struct")
	}

	if reflect.ValueOf(dst).Kind() != reflect.Pointer || reflect.ValueOf(dst).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a pointer to a struct")
	}

	destVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.FieldByName(srcType.Field(i).Name)
		dstField := destVal.FieldByName(srcType.Field(i).Name)

		if dstField == (reflect.Value{}) {
			return fmt.Errorf("field %s is not present on dst struct", srcType.Field(i).Name)
		}

		if srcField.Kind() == reflect.Struct {
			srcNestedStruct := srcField.Interface()
			dstNestedStruct := reflect.New(dstField.Type()).Interface()

			if err := enc.buildDstStruct(srcNestedStruct, dstNestedStruct, op); err != nil {
				return err
			}

			dstField.Set(reflect.ValueOf(dstNestedStruct).Elem())
			continue
		}

		if tagOp, _ := srcType.Field(i).Tag.Lookup(SQIDS_TAG); encoderOperation(tagOp) == op {
			if err := enc.processField(srcField, dstField, encoderOperation(tagOp)); err != nil {
				return fmt.Errorf("error while processing field %s: %w", srcType.Field(i).Name, err)
			}
			continue
		}

		if !srcField.Type().AssignableTo(dstField.Type()) {
			fieldName := srcType.Field(i).Name
			srcTypeName := srcField.Type().Name()
			dstTypeName := dstField.Type().Name()

			return fmt.Errorf("field %s(%s) is not assignable to %s(%s)", fieldName, srcTypeName, fieldName, dstTypeName)
		}

		dstField.Set(srcField)
	}

	return nil
}

func (enc sqidsencoder) processField(srcField, dstField reflect.Value, op encoderOperation) error {
	switch op {
	case ENCODE:
		if srcField.Kind() != reflect.Uint64 {
			return fmt.Errorf("field is not uint64")
		}
		return enc.encodeField(dstField, srcField.Uint())

	case DECODE:
		if srcField.Kind() != reflect.String {
			return fmt.Errorf("field is not string")
		}
		return enc.decodeField(dstField, srcField.String())
	default:
		return fmt.Errorf("unknown operation: %s", op)
	}
}

func (enc sqidsencoder) encodeField(field reflect.Value, id uint64) error {
	encodedID, err := enc.sqids.Encode([]uint64{id})

	if err != nil {
		return err
	}

	if !reflect.TypeOf(encodedID).AssignableTo(field.Type()) {
		return fmt.Errorf("type uint64 is not assignable to %s", field.Type().Name())
	}

	field.SetString(encodedID)
	return nil
}

func (enc sqidsencoder) decodeField(field reflect.Value, id string) error {
	decodedID := enc.sqids.Decode(id)[0]

	if !reflect.TypeOf(decodedID).AssignableTo(field.Type()) {
		return fmt.Errorf("type uint64 is not assignable to %s", field.Type().Name())
	}

	field.SetUint(decodedID)
	return nil
}
