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
	destVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		dstField := destVal.FieldByName(srcType.Field(i).Name)
		srcField := srcVal.FieldByName(srcType.Field(i).Name)

		if tagOp, _ := srcType.Field(i).Tag.Lookup(SQIDS_TAG); encoderOperation(tagOp) == op {
			if err := enc.processField(srcField, dstField, encoderOperation(tagOp)); err != nil {
				return err
			}
			continue
		}

		if !srcField.Type().AssignableTo(dstField.Type()) {
			fieldName := srcType.Field(i).Name
			srcTypeName := srcField.Type().Name()
			dstTypeName := dstField.Type().Name()

			return typeAssigmentError(fieldName, srcTypeName, dstTypeName)
		}

		dstField.Set(srcField)
	}

	return nil
}

func (enc sqidsencoder) processField(srcField, dstField reflect.Value, op encoderOperation) error {
	switch op {
	case ENCODE:
		return enc.encodeField(dstField, srcField.Int())
	case DECODE:
		return enc.decodeField(dstField, srcField.String())
	default:
		return fmt.Errorf("unknown operation: %s", op)
	}
}

func (enc sqidsencoder) encodeField(field reflect.Value, id int64) error {
	encodedID, err := enc.sqids.Encode([]uint64{uint64(id)})

	if err != nil {
		return err
	}

	field.SetString(encodedID)
	return nil
}

func (enc sqidsencoder) decodeField(field reflect.Value, id string) error {
	decodedID := enc.sqids.Decode(id)[0]

	field.SetInt(int64(decodedID))
	return nil
}
