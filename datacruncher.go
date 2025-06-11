package datacruncher

import (
	"fmt"
	"errors"
	"encoding/binary"
	"bytes"
	"reflect"
	"math"
	"strings"
)

func Serialise(data any) ([]byte, error) {
	switch value := data.(type) {
	case bool:
		if value {
			return []byte{1}, nil
		} else {
			return []byte{0}, nil
		}

	case uint8:
		return []byte{value}, nil
	case int8:
		return Serialise(uint8(value))

	case uint16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, value)
		return buf, nil
	case int16:
		return Serialise(uint16(value))

	case uint32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, value)
		return buf, nil
	case int32:
		return Serialise(uint32(value))

	case uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, value)
		return buf, nil
	case int64:
		return Serialise(uint64(value))

	case float32:
		return Serialise(math.Float32bits(value))
	case float64:
		return Serialise(math.Float64bits(value))

	case uint:
		return Serialise(uint64(value))
	case int:
		return Serialise(uint64(value))

	case string:
		return Serialise([]byte(value))
	}

	value := reflect.ValueOf(data)
	kind := value.Kind()

	switch kind {
	case reflect.Array, reflect.Slice:
		var buf bytes.Buffer

		for i := 0; i < value.Len(); i++ {
			element, e := Serialise(value.Index(i).Interface())
			if e != nil {
				return []byte{}, e
			}

			if kind != reflect.Array {
				if len(element) == 0 || element[0] == '\x17' || element[0] == '\x19' {
					buf.WriteByte('\x17')
				}
			}

			buf.Write(element)
		}

		if kind != reflect.Array {
			buf.WriteByte('\x19')
		}

		return buf.Bytes(), nil

	case reflect.Map:
		var buf bytes.Buffer

		keys := value.MapKeys()

		for i := 0; i < value.Len(); i++ {
			key := keys[i]

			mapContent, e := Serialise(key.Interface())
			if e != nil {
				return []byte{}, e
			}

			if len(mapContent) == 0 || mapContent[0] == '\x17' || mapContent[0] == '\x19' {
				buf.WriteByte('\x17')
			}

			buf.Write(mapContent)

			mapValue := value.MapIndex(key)

			mapContent, e = Serialise(mapValue.Interface())

			buf.Write(mapContent)
		}

		buf.WriteByte('\x19')

		return buf.Bytes(), nil

	case reflect.Struct:
		structType := reflect.TypeOf(data)

		var buf bytes.Buffer

		for i := 0; i < value.NumField(); i++ {
			if !structType.Field(i).IsExported() {
				continue
			}

			element, e := Serialise(value.Field(i).Interface())
			if e != nil {
				return []byte{}, e
			}

			buf.Write(element)
		}

		return buf.Bytes(), nil

	case reflect.Pointer:
		if value.IsZero() {
			return []byte{'\x19'}, nil
		}

		content, e := Serialise(reflect.Indirect(value).Interface())
		if e != nil {
			return []byte{}, e
		}

		if len(content) == 0 || content[0] == '\x17' || content[0] == '\x19' {
			content = append([]byte{'\x17'}, content...)
		}
		
		return content, nil
	}

	return []byte{}, errors.New("serialise type not supported")
}

func deserialiseSegment(data []byte, value reflect.Value, i *int) error {
	kind := value.Kind()

	switch kind {
	case reflect.Bool:
		if *i >= len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetBool(data[*i] != '\x00')

		*i++

	case reflect.Int8:
		if *i >= len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetInt(int64(data[*i]))

		*i++
	
	case reflect.Uint8:
		if *i >= len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetUint(uint64(data[*i]))

		*i++

	case reflect.Int16:
		if *i + 2 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetInt(int64(binary.BigEndian.Uint16(data[*i : *i + 2])))

		*i += 2

	case reflect.Uint16:
		if *i + 2 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetUint(uint64(binary.BigEndian.Uint16(data[*i : *i + 2])))

		*i += 2

	case reflect.Int32:
		if *i + 4 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetInt(int64(binary.BigEndian.Uint32(data[*i : *i + 4])))

		*i += 4

	case reflect.Uint32:
		if *i + 4 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetUint(uint64(binary.BigEndian.Uint32(data[*i : *i + 4])))

		*i += 4

	case reflect.Int64, reflect.Int:
		if *i + 8 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetInt(int64(binary.BigEndian.Uint64(data[*i : *i + 8])))

		*i += 8

	case reflect.Uint64, reflect.Uint:
		if *i + 8 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetUint(uint64(binary.BigEndian.Uint64(data[*i : *i + 8])))

		*i += 8
	
	case reflect.Float32:
		if *i + 4 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetFloat(float64(math.Float32frombits(binary.BigEndian.Uint32(data[*i : *i + 4]))))

		*i += 4

	case reflect.Float64:
		if *i + 8 > len(data) {
			return errors.New("deserialise data too short")
		}

		value.SetFloat(float64(math.Float64frombits(binary.BigEndian.Uint64(data[*i : *i + 8]))))

		*i += 8

	case reflect.Array, reflect.Slice, reflect.String:
		if *i >= len(data) && kind != reflect.Array {
			return errors.New("deserialise data too short")
		}

		if kind == reflect.Slice {
			value.Set(reflect.MakeSlice(value.Type(), 0, 0))
		}

		var builder strings.Builder

		added := 0

		for *i < len(data) || kind == reflect.Array {
			if kind != reflect.Array {
				if data[*i] == '\x19' {
					if kind == reflect.String {
						value.SetString(builder.String())
					}
					*i++
					return nil
				}

				if data[*i] == '\x17' {
					*i++
				}
			}

			switch kind {
			case reflect.Array:
				if added == value.Len() {
					return nil
				}
				e := deserialiseSegment(data, value.Index(added), i)
				if e != nil {
					return e
				}
				added++
			case reflect.Slice:
				newValue := reflect.Indirect(reflect.New(value.Type().Elem()))

				e := deserialiseSegment(data, newValue, i)
				if e != nil {
					return e
				}
				value.Set(reflect.Append(value, newValue))
			case reflect.String:
				builder.WriteByte(data[*i])
				*i++
			}
		}

		return errors.New("deserialise sequence not terminated")

	case reflect.Struct:
		structType := value.Type()

		for f := 0; f < value.NumField(); f++ {
			if !structType.Field(f).IsExported() {
				continue
			}

			e := deserialiseSegment(data, value.Field(f), i)
			if e != nil {
				return e
			}
		}

	case reflect.Map:
		if *i >= len(data) {
			return errors.New("deserialise data too short")
		}

		value.Set(reflect.MakeMap(value.Type()))

		for *i < len(data) {
			if data[*i] == '\x19' {
				*i++
				return nil
			}

			if data[*i] == '\x17' {
				*i++
			}

			key := reflect.Indirect(reflect.New(value.Type().Key()))
			e := deserialiseSegment(data, key, i)
			if e != nil {
				return e
			}

			mapValue := reflect.Indirect(reflect.New(value.Type().Elem()))
			e = deserialiseSegment(data, mapValue, i)

			value.SetMapIndex(key, mapValue)
		}

		return errors.New("deserialise map sequence not terminated")

	case reflect.Pointer:
		if *i >= len(data) {
			return errors.New("deserialise data too short")
		}

		if data[0] == '\x19' {
			value.Set(reflect.Zero(value.Type()))
			*i++
			return nil
		}

		if data[0] == '\x17' {
			*i++
		}

		value.Set(reflect.New(value.Type().Elem()))

		deserialiseSegment(data, reflect.Indirect(value), i)

	default:
		return fmt.Errorf("deserialise container kind not supported: %v", kind)
	}

	return nil
}

func Deserialise(data []byte, container any) error {
	containerPointer := reflect.ValueOf(container)

	if containerPointer.Kind() != reflect.Pointer {
		return errors.New("deserialise container is not a pointer")
	}

	i := 0

	return deserialiseSegment(data, reflect.Indirect(containerPointer), &i)
}
