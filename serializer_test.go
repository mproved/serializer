package serializer

import (
	"reflect"
	"testing"
)

func DoTest(t *testing.T, value any) {
	encoded := Encode(value)
	decodedValue := Decode(encoded)

	if !reflect.DeepEqual(value, decodedValue) {
		t.Fatalf(`serializing failed: given=%v got=%v`, value, decodedValue)
	}
}

func TestSerializerBool(t *testing.T) {
	DoTest(t, bool(true))
}

func TestSerializerInt(t *testing.T) {
	DoTest(t, int(-1234567890))
}

func TestSerializerInt8(t *testing.T) {
	DoTest(t, int8(-123))
}

func TestSerializerInt16(t *testing.T) {
	DoTest(t, int16(-12345))
}

func TestSerializerInt32(t *testing.T) {
	DoTest(t, int32(-1234567890))
}

func TestSerializerInt64(t *testing.T) {
	DoTest(t, int64(-1234567890))
}

func TestSerializerUint(t *testing.T) {
	DoTest(t, uint(1234567890))
}

func TestSerializerUint8(t *testing.T) {
	DoTest(t, uint8(123))
}

func TestSerializerUint16(t *testing.T) {
	DoTest(t, uint16(12345))
}

func TestSerializerUint32(t *testing.T) {
	DoTest(t, uint32(1234567890))
}

func TestSerializerUint64(t *testing.T) {
	DoTest(t, uint64(1234567890))
}

func TestSerializerFloat32(t *testing.T) {
	DoTest(t, float32(1234567890.12345))
}

func TestSerializerFloat64(t *testing.T) {
	DoTest(t, float64(1234567890.12345))
}

func TestSerializerInt32Array(t *testing.T) {
	DoTest(t, [5]int32{1, 2, 3, 4, 5})
}

func TestSerializerMapInt32MapInt32Int32(t *testing.T) {
	m := map[int32]map[int32]int32{
		5: {
			3: 6,
		},
	}

	DoTest(t, m)
}

func TestSerializerFixed5ArrayFixed5ArrayInt32(t *testing.T) {
	DoTest(t, [5][5]int32{
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
	})
}

type Foo struct {
	Bar int
}

func TestSerializerFoo(t *testing.T) {
	RegisterType(Foo{}, 50)

	DoTest(t, Foo{})
}
