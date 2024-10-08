package sqidsencoder

import (
	"errors"
	"fmt"
	"reflect"
)

type encoderOperation string

const (
	SQIDS_TAG                  = "sqids"
	ENCODE    encoderOperation = "encode"
	DECODE    encoderOperation = "decode"
)

var (
	ErrInvalidInput = errors.New("validation error")
	ErrType         = errors.New("type error")
	ErrInvalidID    = errors.New("invalid ID")
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
	return enc.processStructFields(src, dst, ENCODE)
}

func (enc sqidsencoder) Decode(src any, dst any) error {
	return enc.processStructFields(src, dst, DECODE)
}

func (enc sqidsencoder) processStructFields(src any, dst any, op encoderOperation) error {
	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)

	if srcVal.Kind() != reflect.Struct {
		return fmt.Errorf("%w: src must be a struct", ErrInvalidInput)
	}

	if reflect.ValueOf(dst).Kind() != reflect.Pointer || reflect.ValueOf(dst).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%w: dst must be a pointer to a struct", ErrInvalidInput)
	}

	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.FieldByName(srcType.Field(i).Name)
		dstField := dstVal.FieldByName(srcType.Field(i).Name)

		if dstField == (reflect.Value{}) {
			return fmt.Errorf("%w: field %s is not present on dst struct", ErrInvalidInput, srcType.Field(i).Name)
		}

		tagOp, hasTag := srcType.Field(i).Tag.Lookup(SQIDS_TAG)
		if hasTag && encoderOperation(tagOp) == op {
			if err := enc.processField(srcField, dstField, op); err != nil {
				return fmt.Errorf("error while processing field %s: %w", srcType.Field(i).Name, err)
			}
			continue
		}

		if err := assignField(srcField, dstField); err != nil {
			return fmt.Errorf("error while processing field %s: %w", srcType.Field(i).Name, err)
		}
	}

	return nil
}

func (enc sqidsencoder) processField(srcField, dstField reflect.Value, op encoderOperation) error {
	switch srcField.Kind() {
	case reflect.Slice:
		return enc.processSlice(srcField, dstField, op)
	case reflect.Struct:
		return enc.processStruct(srcField, dstField, op)

	default:
		return enc.processPrimitive(srcField, dstField, op)
	}
}

func (enc sqidsencoder) processStruct(srcStruct, dstStruct reflect.Value, op encoderOperation) error {
	srcNestedStruct := srcStruct.Interface()
	dstNestedStruct := reflect.New(dstStruct.Type()).Interface()

	if err := enc.processStructFields(srcNestedStruct, dstNestedStruct, op); err != nil {
		return err
	}

	dstStruct.Set(reflect.ValueOf(dstNestedStruct).Elem())

	return nil
}

func (enc sqidsencoder) processSlice(srcSliceField, dstSliceField reflect.Value, op encoderOperation) error {
	switch op {
	case ENCODE:
		return enc.encodeSlice(srcSliceField, dstSliceField)

	case DECODE:
		return enc.decodeSlice(srcSliceField, dstSliceField)

	default:
		return fmt.Errorf("unknown operation: %s", op)
	}
}

func (enc sqidsencoder) processPrimitive(srcField, dstField reflect.Value, op encoderOperation) error {
	switch op {
	case ENCODE:
		if srcField.Kind() != reflect.Uint64 {
			return fmt.Errorf("%w: field is not uint64", ErrType)
		}
		return enc.encodeField(dstField, srcField.Uint())

	case DECODE:
		if srcField.Kind() != reflect.String {
			return fmt.Errorf("%w: field is not string", ErrType)
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

	return assignField(reflect.ValueOf(encodedID), field)
}

func (enc sqidsencoder) encodeSlice(srcField, dstField reflect.Value) error {
	encodedSlice := reflect.MakeSlice(dstField.Type(), srcField.Cap(), srcField.Cap())

	if srcField.Type().Elem().Kind() == reflect.Uint64 {
		for i := 0; i < srcField.Len(); i++ {
			if err := enc.encodeField(encodedSlice.Index(i), srcField.Index(i).Uint()); err != nil {
				return err
			}
		}
		dstField.Set(encodedSlice)
	}

	if srcField.Type().Elem().Kind() == reflect.Struct {
		for i := 0; i < srcField.Len(); i++ {
			if err := enc.processStruct(srcField.Index(i), encodedSlice.Index(i), ENCODE); err != nil {
				return err
			}
		}
		dstField.Set(encodedSlice)
	}

	return nil
}

func (enc sqidsencoder) decodeField(field reflect.Value, id string) error {
	res := enc.sqids.Decode(id)
	if len(res) == 0 {
		return fmt.Errorf("%w: %s", ErrInvalidID, id)
	}
	decodedID := res[0]

	return assignField(reflect.ValueOf(decodedID), field)
}

func (enc sqidsencoder) decodeSlice(srcField, dstField reflect.Value) error {
	decodedSlice := reflect.MakeSlice(dstField.Type(), srcField.Cap(), srcField.Cap())

	if srcField.Type().Elem().Kind() == reflect.String {
		for i := 0; i < srcField.Len(); i++ {
			if err := enc.decodeField(decodedSlice.Index(i), srcField.Index(i).String()); err != nil {
				return err
			}
		}
		dstField.Set(decodedSlice)
	}

	if srcField.Type().Elem().Kind() == reflect.Struct {
		for i := 0; i < srcField.Len(); i++ {
			if err := enc.processStruct(srcField.Index(i), decodedSlice.Index(i), DECODE); err != nil {
				return err
			}
		}
		dstField.Set(decodedSlice)
	}

	return nil
}

func assignField(srcField, dstField reflect.Value) error {
	if !srcField.Type().AssignableTo(dstField.Type()) {
		var hint string

		if srcField.Type().Kind() == reflect.Uint64 && dstField.Type().Kind() == reflect.String {
			hint = "you may have forgotten to add the tag `sqids:\"encode\"`"
		}

		if srcField.Type().Kind() == reflect.String && dstField.Type().Kind() == reflect.Uint64 {
			hint = "you may have forgotten to add the tag `sqids:\"decode\"`"
		}

		return fmt.Errorf(
			"%w: src type %s is not assignable to dst type %s. %s",
			ErrType,
			srcField.Type().Name(),
			dstField.Type().Name(),
			hint,
		)
	}

	dstField.Set(srcField)
	return nil
}
