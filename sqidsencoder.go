package sqidsencoder

import (
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
		fieldName := srcType.Field(i).Name

		if op, ok := srcType.Field(i).Tag.Lookup("sqids"); ok && op == "encode" {
			encodedID, err := enc.sqids.Encode([]uint64{uint64(srcVal.Field(i).Int())})

			if err != nil {
				return err
			}

			destVal.FieldByName(fieldName).SetString(encodedID)
			continue
		}

		destVal.FieldByName(fieldName).Set(srcVal.FieldByName(fieldName))
	}

	return nil
}
