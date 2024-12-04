package packet_test

import (
	"bytes"
	"fmt"
	"testing"

	pk "github.com/Tnze/go-mc/net/packet"
)

func ExampleAry_WriteTo() {
	data := []pk.Int{0, 1, 2, 3, 4, 5, 6}
	// Len is completely ignored by WriteTo method.
	// The length is inferred from the length of Ary.
	pk.Marshal(
		0x00,
		pk.Ary[pk.VarInt]{
			Ary: data,
		},
	)
}

func ExampleAry_ReadFrom() {
	var data []pk.String

	var p pk.Packet // = conn.ReadPacket()
	if err := p.Scan(
		pk.Ary[pk.VarInt]{ // then decode Ary according to length
			Ary: &data,
		},
	); err != nil {
		panic(err)
	}
}

func TestAry_ReadFrom(t *testing.T) {
	var ary []pk.String
	bin := []byte{
		0, 0, 0, 2,
		4, 'T', 'n', 'z', 'e',
		0,
	}
	data := pk.Ary[pk.Int]{Ary: &ary}
	if _, err := data.ReadFrom(bytes.NewReader(bin)); err != nil {
		t.Fatal(err)
	}
	if len(ary) != 2 {
		t.Fatalf("length not match: %d != %d", len(ary), 2)
	}
	for i, v := range []string{"Tnze", ""} {
		if string(ary[i]) != v {
			t.Errorf("want %q, get %q", v, ary[i])
		}
	}
}

func TestAry_WriteTo(t *testing.T) {
	var buf bytes.Buffer
	want := []byte{
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03,
	}
	for _, item := range [...]pk.FieldEncoder{
		pk.Ary[pk.Int]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.Int]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.Long]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.VarInt]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.VarLong]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.Int]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.Long]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.VarInt]{Ary: []pk.Int{1, 2, 3}},
		pk.Ary[pk.VarLong]{Ary: []pk.Int{1, 2, 3}},
	} {
		_, err := item.WriteTo(&buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes()[buf.Len()-3*4:], want) {
			t.Fatalf("Ary encoding error: got %#v, want %#v", buf.Bytes(), want)
		}
		buf.Reset()
	}
}

func TestAry_WriteTo_pointer(t *testing.T) {
	var buf bytes.Buffer
	want := []byte{
		0x00, 0x00, 0x00, 0x03,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03,
	}
	data := pk.Ary[pk.Int]{Ary: &[]pk.Int{1, 2, 3}}

	_, err := data.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("Ary encoding error: got %#v, want %#v", buf.Bytes(), want)
	}
}

func ExampleOpt_ReadFrom() {
	var has pk.Boolean

	var data pk.String
	p1 := pk.Packet{Data: []byte{
		0x01,                  // pk.Boolean(true)
		4, 'T', 'n', 'z', 'e', // pk.String
	}}
	if err := p1.Scan(
		&has,
		pk.Opt{
			Has: &has, Field: &data,
		},
	); err != nil {
		panic(err)
	}
	fmt.Println(data)

	var data2 pk.String = "WILL NOT BE READ, WILL NOT BE COVERED"
	p2 := pk.Packet{Data: []byte{
		0x00, // pk.Boolean(false)
		// empty
	}}
	if err := p2.Scan(
		&has,
		pk.Opt{Has: &has, Field: &data2},
	); err != nil {
		panic(err)
	}
	fmt.Println(data2)

	// Output:
	// Tnze
	// WILL NOT BE READ, WILL NOT BE COVERED
}

// As an example, we define this packet as this:
// +------+-----------------+----------------------------------+
// | Name | Type            | Notes                            |
// +------+-----------------+----------------------------------+
// | Flag | Unsigned Byte   | Odd if the following is present. |
// +------+-----------------+----------------------------------+
// | User | Optional String | The player's name.               |
// +------+-----------------+----------------------------------+
// So we need a function to decide if the User field is present.
func ExampleOpt_ReadFrom_func() {
	var flag pk.Byte
	var data pk.String
	p := pk.Packet{Data: []byte{
		0b_0010_0011,          // pk.Byte(flag)
		4, 'T', 'n', 'z', 'e', // pk.String
	}}
	if err := p.Scan(
		&flag,
		pk.Opt{
			Has: func() bool {
				return flag&1 != 0
			},
			Field: &data,
		},
	); err != nil {
		panic(err)
	}
	fmt.Println(data)

	// Output: Tnze
}

func ExampleTuple_ReadFrom() {
	// When you need to read an "Optional Array of X":
	var has pk.Boolean
	var ary []pk.String

	var p pk.Packet // = conn.ReadPacket()
	if err := p.Scan(
		&has,
		pk.Opt{
			Has: &has,
			Field: pk.Tuple{
				pk.Ary[pk.Int]{Ary: &ary},
			},
		},
	); err != nil {
		panic(err)
	}
}

// As an example, we define this packet as this:
// +------+-----------------+-----------------------------------+
// | Name | Type            | Notes                             |
// +------+-----------------+-----------------------------------+
// | Has  | Boolean         | True if the following is present. |
// +------+-----------------+-----------------------------------+
// | User | Optional String | The player's name.                |
// +------+-----------------+-----------------------------------+
// So we need a function to decide if the User field is present.
func ExampleOption_ReadFrom_func() {
	p1 := pk.Packet{Data: []byte{
		0x01,                  // pk.Boolean(true)
		4, 'T', 'n', 'z', 'e', // pk.String("Tnze")
	}}
	p2 := pk.Packet{Data: []byte{
		0x00, // pk.Boolean(false)
		// empty
	}}

	var User1, User2 pk.Option[pk.String, *pk.String]
	if err := p1.Scan(&User1); err != nil {
		panic(err)
	}
	if err := p2.Scan(&User2); err != nil {
		panic(err)
	}

	fmt.Println(User1.Has, User1.Val)
	fmt.Println(User2.Has, User2.Val)

	// Output:
	// true Tnze
	// false
}
