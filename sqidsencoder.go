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

	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.FieldByName(srcType.Field(i).Name)
		dstField := dstVal.FieldByName(srcType.Field(i).Name)

		if dstField == (reflect.Value{}) {
			return fmt.Errorf("field %s is not present on dst struct", srcType.Field(i).Name)
		}

		tagOp, hasTag := srcType.Field(i).Tag.Lookup(SQIDS_TAG)
		if hasTag && encoderOperation(tagOp) == op {
			if err := enc.processTaggedField(srcField, dstField, op); err != nil {
				return fmt.Errorf("error while processing tagged field %s: %w", srcType.Field(i).Name, err)
			}
			continue
		}

		if srcField.Kind() == reflect.Struct {
			if err := enc.processStruct(srcField, dstField, op); err != nil {
				return fmt.Errorf("error while processing struct field %s: %w", srcType.Field(i).Name, err)
			}
			continue
		}

		if err := assignField(srcField, dstField, srcType.Field(i).Name); err != nil {
			return err
		}
	}

	return nil
}

func (enc sqidsencoder) processStruct(srcStruct, dstStruct reflect.Value, op encoderOperation) error {
	srcNestedStruct := srcStruct.Interface()
	dstNestedStruct := reflect.New(dstStruct.Type()).Interface()

	if err := enc.buildDstStruct(srcNestedStruct, dstNestedStruct, op); err != nil {
		return err
	}

	dstStruct.Set(reflect.ValueOf(dstNestedStruct).Elem())

	return nil
}

func (enc sqidsencoder) processTaggedField(srcField, dstField reflect.Value, op encoderOperation) error {
	if srcField.Kind() == reflect.Slice {
		return enc.processSlice(srcField, dstField, op)
	}

	return enc.processField(srcField, dstField, op)
}

func (enc sqidsencoder) processSlice(srcSliceField, dstSlicefield reflect.Value, op encoderOperation) error {
	switch op {
	case ENCODE:
		ids, ok := srcSliceField.Interface().([]uint64)
		if !ok {
			return fmt.Errorf("field is not []uint64")
		}
		return enc.encodeSlice(dstSlicefield, ids)

	case DECODE:
		ids, ok := srcSliceField.Interface().([]string)
		if !ok {
			return fmt.Errorf("field is not []string")
		}

		return enc.decodeSlice(dstSlicefield, ids)

	default:
		return fmt.Errorf("unknown operation: %s", op)
	}
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

func (enc sqidsencoder) encodeSlice(field reflect.Value, ids []uint64) error {
	encodedSlice := make([]string, len(ids))

	if !reflect.TypeOf(encodedSlice).AssignableTo(field.Type()) {
		return fmt.Errorf("type []uint64 is not assignable to %s", field.Type().Name())
	}

	for i := range ids {
		encodedID, err := enc.sqids.Encode([]uint64{ids[i]})

		if err != nil {
			return err
		}

		encodedSlice[i] = encodedID
	}

	field.Set(reflect.ValueOf(encodedSlice))
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

func (enc sqidsencoder) decodeSlice(field reflect.Value, ids []string) error {
	decodedSlice := make([]uint64, len(ids))

	if !reflect.TypeOf(decodedSlice).AssignableTo(field.Type()) {
		return fmt.Errorf("type []uint64 is not assignable to %s", field.Type().Name())
	}

	for i := range ids {
		encodedID := enc.sqids.Decode(ids[i])[0]

		decodedSlice[i] = encodedID
	}

	field.Set(reflect.ValueOf(decodedSlice))
	return nil
}

func assignField(srcField, dstField reflect.Value, fieldName string) error {
	if !srcField.Type().AssignableTo(dstField.Type()) {
		return fmt.Errorf(
			"field src.%s(%s) is not assignable to dst.%s(%s)",
			fieldName,
			srcField.Type().Name(),
			fieldName,
			dstField.Type().Name(),
		)
	}

	dstField.Set(srcField)
	return nil
}
