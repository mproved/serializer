package serializer

import (
	"hash/fnv"
	"reflect"

	"errors"

	"github.com/mproved/little_endian_buffer"
)

type TypeId uint16

const (
	TypeIdUnknown TypeId = 0
	TypeIdBool    TypeId = 1
	TypeIdInt     TypeId = 2
	TypeIdInt8    TypeId = 3
	TypeIdInt16   TypeId = 4
	TypeIdInt32   TypeId = 5
	TypeIdInt64   TypeId = 6
	TypeIdUint    TypeId = 7
	TypeIdUint8   TypeId = 8
	TypeIdUint16  TypeId = 9
	TypeIdUint32  TypeId = 10
	TypeIdUint64  TypeId = 11
	TypeIdFloat32 TypeId = 12
	TypeIdFloat64 TypeId = 13
	TypeIdArray   TypeId = 14
	TypeIdMap     TypeId = 15
	TypeIdSlice   TypeId = 16
	TypeIdString  TypeId = 17
)

var ErrCannotEncode = errors.New("cannot encode")

var typeToTypeIdMapping = map[reflect.Type]TypeId{}
var typeIdToTypeMapping = map[TypeId]reflect.Type{}

var typeIdNameHashMapping = map[TypeId]map[string]uint32{}
var typeIdHashNameMapping = map[TypeId]map[uint32]string{}

func init() {
	RegisterType(false, TypeIdBool)
	RegisterType(int(0), TypeIdInt)
	RegisterType(int8(0), TypeIdInt8)
	RegisterType(int16(0), TypeIdInt16)
	RegisterType(int32(0), TypeIdInt32)
	RegisterType(int64(0), TypeIdInt64)
	RegisterType(uint(0), TypeIdUint)
	RegisterType(uint8(0), TypeIdUint8)
	RegisterType(uint16(0), TypeIdUint16)
	RegisterType(uint32(0), TypeIdUint32)
	RegisterType(uint64(0), TypeIdUint64)
	RegisterType(float32(0), TypeIdFloat32)
	RegisterType(float64(0), TypeIdFloat64)
	RegisterType("", TypeIdString)
}

func RegisterType(value any, typeId TypeId) {
	typeOf := reflect.TypeOf(value)

	// XXX
	// check for dups
	typeToTypeIdMapping[typeOf] = typeId
	typeIdToTypeMapping[typeId] = typeOf

	if typeOf.Kind() == reflect.Struct {
		typeIdNameHashMapping[typeId] = make(map[string]uint32)
		typeIdHashNameMapping[typeId] = make(map[uint32]string)

		valueOf := reflect.ValueOf(value)

		for i := 0; i < valueOf.NumField(); i++ {
			fieldInfo := typeOf.Field(i)

			h := fnv.New32a()
			h.Write([]byte(fieldInfo.Name))
			nameHash := h.Sum32()

			// XXX
			// check for collusions
			typeIdNameHashMapping[typeId][fieldInfo.Name] = nameHash
			typeIdHashNameMapping[typeId][nameHash] = fieldInfo.Name
		}
	}
}

func writeTypeIdFixed(b *little_endian_buffer.Buffer, typeOf reflect.Type) {
	switch typeOf.Kind() {
	case reflect.Bool:
		b.WriteUint16(uint16(TypeIdBool))
	case reflect.Int:
		b.WriteUint16(uint16(TypeIdInt))
	case reflect.Int8:
		b.WriteUint16(uint16(TypeIdInt8))
	case reflect.Int16:
		b.WriteUint16(uint16(TypeIdInt16))
	case reflect.Int32:
		b.WriteUint16(uint16(TypeIdInt32))
	case reflect.Int64:
		b.WriteUint16(uint16(TypeIdInt64))
	case reflect.Uint:
		b.WriteUint16(uint16(TypeIdUint))
	case reflect.Uint8:
		b.WriteUint16(uint16(TypeIdUint8))
	case reflect.Uint16:
		b.WriteUint16(uint16(TypeIdUint16))
	case reflect.Uint32:
		b.WriteUint16(uint16(TypeIdUint32))
	case reflect.Uint64:
		b.WriteUint16(uint16(TypeIdUint64))
	case reflect.Float32:
		b.WriteUint16(uint16(TypeIdFloat32))
	case reflect.Float64:
		b.WriteUint16(uint16(TypeIdFloat64))
	case reflect.Array:
		b.WriteUint16(uint16(TypeIdArray))
	case reflect.Map:
		b.WriteUint16(uint16(TypeIdMap))
	case reflect.Slice:
		b.WriteUint16(uint16(TypeIdSlice))
	case reflect.String:
		b.WriteUint16(uint16(TypeIdString))
	default:
		b.WriteUint16(uint16(typeToTypeIdMapping[typeOf]))
	}
}

func Decode(b []byte) any {
	buf := little_endian_buffer.BufferFromBytes(b)

	return decodeInternal(buf, []TypeId{}, TypeIdUnknown)
}

func decodeInternal(buf *little_endian_buffer.Buffer, typeIds []TypeId, typeId TypeId) any {
	if buf.LeftToRead() == 0 {
		return nil
	}

	if typeId == TypeIdUnknown {
		typeId = TypeId(buf.ReadUint16())
	}

	switch typeId {

	case TypeIdBool:

		return buf.ReadBool()

	case TypeIdInt:

		size := buf.ReadUint8()

		switch size {
		case 1:
			return int(buf.ReadInt8())
		case 2:
			return int(buf.ReadInt16())
		case 4:
			return int(buf.ReadInt32())
		case 8:
			return int(buf.ReadInt64())
		}

	case TypeIdInt8:

		return buf.ReadInt8()

	case TypeIdInt16:

		return buf.ReadInt16()

	case TypeIdInt32:

		return buf.ReadInt32()

	case TypeIdInt64:

		return buf.ReadInt64()

	case TypeIdUint:

		size := buf.ReadUint8()

		switch size {
		case 1:
			return uint(buf.ReadUint8())
		case 2:
			return uint(buf.ReadUint16())
		case 4:
			return uint(buf.ReadUint32())
		case 8:
			return uint(buf.ReadUint64())
		}

	case TypeIdUint8:

		return buf.ReadUint8()

	case TypeIdUint16:

		return buf.ReadUint16()

	case TypeIdUint32:

		return buf.ReadUint32()

	case TypeIdUint64:

		return buf.ReadUint64()

	case TypeIdFloat32:

		return buf.ReadFloat32()

	case TypeIdFloat64:

		return buf.ReadFloat64()

	case TypeIdString:

		len := buf.ReadUint16()
		return string(buf.ReadBytes(int(len)))

	case TypeIdArray, TypeIdSlice:

		length := int(buf.ReadUint16())

		if len(typeIds) == 0 {
			typesLength := int(buf.ReadUint16())

			for i := 0; i < typesLength; i += 1 {
				typeId := TypeId(buf.ReadUint16())

				typeIds = append(typeIds, typeId)
			}
		}

		items := make(map[int]any)

		for i := 0; i < length; i += 1 {
			items[i] = decodeInternal(buf, typeIds[1:], typeIds[0])
		}

		typeOf := reflect.TypeOf(items[0])

		if typeOf == nil {
			return nil
		}

		var created reflect.Value

		if typeId == TypeIdArray {
			arrayOf := reflect.ArrayOf(length, typeOf)

			// elem because its a pointer
			created = reflect.New(arrayOf).Elem()
		}

		if typeId == TypeIdSlice {
			sliceOf := reflect.SliceOf(typeOf)

			created = reflect.MakeSlice(sliceOf, length, length)
		}

		for i := 0; i < length; i += 1 {
			created.Index(i).Set(reflect.ValueOf(items[i]))
		}

		return created.Interface()

	case TypeIdMap:

		// XXX
		// map should be redone

		// len := int(buf.ReadUint16())

		// itemKeyTypeId := TypeId(buf.ReadUint16())
		// typeOfKey := typeIdToTypeMapping[itemKeyTypeId]

		// itemValueTypeId := TypeId(buf.ReadUint16())
		// typeOfValue := typeIdToTypeMapping[itemValueTypeId]

		// mapOf := reflect.MapOf(typeOfKey, typeOfValue)

		// created := reflect.MakeSlice(mapOf, len, len)

		// for i := 0; i < len; i += 1 {
		// 	itemKey := decodeInternal(buf, []TypeId{}, itemKeyTypeId)
		// 	itemValue := decodeInternal(buf, []TypeId{}, itemValueTypeId)

		// 	created.SetMapIndex(reflect.ValueOf(itemKey), reflect.ValueOf(itemValue))
		// }

		// return created.Interface()

	default:

		_, ok := typeIdToTypeMapping[typeId]

		if !ok {
			return nil
		}

		numberOfFields := int(buf.ReadUint16())

		typeOf := typeIdToTypeMapping[typeId]

		created := reflect.New(typeOf).Elem()

		for i := 0; i < numberOfFields; i += 1 {
			nameHash := buf.ReadUint32()

			name := typeIdHashNameMapping[typeId][nameHash]

			if name == "" {
				continue
			}

			field := decodeInternal(buf, []TypeId{}, TypeIdUnknown)

			if field == nil {
				continue
			}

			createField := created.FieldByName(name)

			fieldValue := reflect.ValueOf(field)

			fieldType := reflect.TypeOf(field)

			fieldInfo, _ := typeOf.FieldByName(name)

			if fieldType.AssignableTo(fieldInfo.Type) {
				createField.Set(fieldValue)
			} else if fieldType.ConvertibleTo(fieldInfo.Type) {
				createField.Set(fieldValue.Convert(fieldInfo.Type))
			} else {
				if (fieldType.Kind() == reflect.Array ||
					fieldType.Kind() == reflect.Slice) &&
					(fieldInfo.Type.Kind() == reflect.Array ||
						fieldInfo.Type.Kind() == reflect.Slice) {

					if fieldInfo.Type.Kind() == reflect.Slice {
						createField.Grow(fieldValue.Len())
						createField.SetLen(fieldValue.Len())
					}

					for j := 0; j < fieldValue.Len(); j += 1 {

						t := fieldType.Elem()
						ft := fieldInfo.Type.Elem()

						if t.AssignableTo(ft) {
							createField.Index(j).Set(fieldValue.Index(j))
						} else if t.ConvertibleTo(ft) {
							createField.Index(j).Set(fieldValue.Index(j).Convert(ft))
						}

					}
				}
			}
		}

		return created.Interface()

	}

	return nil
}

func Encode(value any) []byte {
	buf := little_endian_buffer.BufferFromBytes([]byte{})
	encodeInternal(buf, value, true)

	return buf.Bytes()
}

func encodeInternal(buf *little_endian_buffer.Buffer, value any, header bool) {
	typeOf := reflect.TypeOf(value)
	valueOf := reflect.ValueOf(value)
	kind := typeOf.Kind()

	if kind == reflect.Pointer {
		valueOf = valueOf.Elem()
		typeOf = valueOf.Type()
		kind = typeOf.Kind()
	}

	if header {
		writeTypeIdFixed(buf, typeOf)
	}

	switch kind {

	case reflect.Bool:

		buf.WriteBool(valueOf.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		size := typeOf.Size()

		if kind == reflect.Int {
			buf.WriteUint8(uint8(size))
		}

		switch size {
		case 1:
			buf.WriteInt8(int8(valueOf.Int()))
		case 2:
			buf.WriteInt16(int16(valueOf.Int()))
		case 4:
			buf.WriteInt32(int32(valueOf.Int()))
		case 8:
			buf.WriteInt64(int64(valueOf.Int()))
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		size := typeOf.Size()

		if kind == reflect.Uint {
			buf.WriteUint8(uint8(size))
		}

		switch size {
		case 1:
			buf.WriteUint8(uint8(valueOf.Uint()))
		case 2:
			buf.WriteUint16(uint16(valueOf.Uint()))
		case 4:
			buf.WriteUint32(uint32(valueOf.Uint()))
		case 8:
			buf.WriteUint64(uint64(valueOf.Uint()))
		}

	case reflect.Float32:

		buf.WriteFloat32(float32(valueOf.Float()))

	case reflect.Float64:

		buf.WriteFloat64(float64(valueOf.Float()))

	case reflect.String:

		buf.WriteUint16(uint16(valueOf.Len()))
		buf.WriteBytes([]byte(valueOf.String()))

	case reflect.Array, reflect.Slice:

		length := valueOf.Len()
		buf.WriteUint16(uint16(length))

		if header {
			itemTypeOf := typeOf.Elem()

			types := []reflect.Type{}

			types = append(types, itemTypeOf)

			for {
				if itemTypeOf.Kind() != reflect.Array &&
					itemTypeOf.Kind() != reflect.Slice &&
					itemTypeOf.Kind() != reflect.Map {
					break
				}

				itemTypeOf = itemTypeOf.Elem()

				types = append(types, itemTypeOf)
			}

			buf.WriteUint16(uint16(len(types)))

			for _, typeOf := range types {
				writeTypeIdFixed(buf, typeOf)
			}
		}

		for i := 0; i < length; i += 1 {
			item := valueOf.Index(i).Interface()

			encodeInternal(buf, item, false)
		}

	case reflect.Map:

		// XXX
		// map should be redone

		// length := valueOf.Len()
		// buf.WriteUint16(uint16(length))

		// if header {
		// 	itemKeyTypeOf := typeOf.Key()

		// 	keyTypes := []reflect.Type{}

		// 	keyTypes = append(keyTypes, itemKeyTypeOf)

		// 	for {
		// 		if itemKeyTypeOf.Kind() != reflect.Array &&
		// 			itemKeyTypeOf.Kind() != reflect.Slice &&
		// 			itemKeyTypeOf.Kind() != reflect.Map {
		// 			break
		// 		}

		// 		itemKeyTypeOf = itemKeyTypeOf.Elem()

		// 		keyTypes = append(keyTypes, itemKeyTypeOf)
		// 	}

		// 	buf.WriteUint16(uint16(len(keyTypes)))

		// 	for _, typeOf := range keyTypes {
		// 		writeTypeIdFixed(buf, typeOf)
		// 	}

		// 	itemValueTypeOf := typeOf.Elem()

		// 	valueTypes := []reflect.Type{}

		// 	valueTypes = append(valueTypes, itemValueTypeOf)

		// 	for {
		// 		if itemValueTypeOf.Kind() != reflect.Array &&
		// 			itemValueTypeOf.Kind() != reflect.Slice &&
		// 			itemValueTypeOf.Kind() != reflect.Map {
		// 			break
		// 		}

		// 		itemValueTypeOf = itemValueTypeOf.Elem()

		// 		valueTypes = append(valueTypes, itemValueTypeOf)
		// 	}

		// 	buf.WriteUint16(uint16(len(keyTypes)))

		// 	for _, typeOf := range keyTypes {
		// 		writeTypeIdFixed(buf, typeOf)
		// 	}
		// }

		// if header {
		// 	itemKeyTypeOf := typeOf.Key()
		// 	itemValueTypeOf := typeOf.Elem()

		// 	writeTypeIdFixed(buf, itemKeyTypeOf)
		// 	writeTypeIdFixed(buf, itemValueTypeOf)

		// 	innerValueTypeOf := itemValueTypeOf

		// 	for {
		// 		if innerValueTypeOf.Kind() == reflect.Array ||
		// 			innerValueTypeOf.Kind() == reflect.Slice ||
		// 			innerValueTypeOf.Kind() == reflect.Map {
		// 			break

		// 		}

		// 		innerValueTypeOf = itemValueTypeOf.Elem()

		// 		writeTypeIdFixed(buf, innerValueTypeOf)
		// 	}
		// }

		// iter := valueOf.MapRange()

		// for iter.Next() {
		// 	itemKey := iter.Key().Interface()
		// 	itemValue := iter.Value().Interface()

		// 	encodeInternal(buf, itemKey, false)
		// 	encodeInternal(buf, itemValue, false)
		// }

	case reflect.Struct:

		encodedFields := 0

		savedPointer := buf.Pointer()
		buf.WriteUint16(0)

		for i := 0; i < valueOf.NumField(); i++ {
			fieldValue := valueOf.Field(i)

			if fieldValue.IsZero() {
				continue
			}

			fieldInfo := typeOf.Field(i)

			if !fieldInfo.IsExported() {
				continue
			}

			h := fnv.New32a()
			h.Write([]byte(fieldInfo.Name))
			nameHash := h.Sum32()
			buf.WriteUint32(nameHash)

			field := fieldValue.Interface()
			encodeInternal(buf, field, true)

			encodedFields += 1
		}

		restoredPointer := buf.Pointer()

		buf.SetPointer(savedPointer)
		buf.WriteUint16(uint16(encodedFields))
		buf.SetPointer(restoredPointer)

	}
}
