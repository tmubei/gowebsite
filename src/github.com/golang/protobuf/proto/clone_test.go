// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2011 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto

import (
	"testing"

	proto3pb "github.com/golang/protobuf/proto/proto3_proto"
	pb "github.com/golang/protobuf/proto/test_proto"
)

var cloneTestMessage = &pb.MyMessage{
	Count: Int32(42),
	Name:  String("Dave"),
	Pet:   []string{"bunny", "kitty", "horsey"},
	Inner: &pb.InnerMessage{
		Host:      String("niles"),
		Port:      Int32(9099),
		Connected: Bool(true),
	},
	Others: []*pb.OtherMessage{
		{
			Value: []byte("some bytes"),
		},
	},
	Somegroup: &pb.MyMessage_SomeGroup{
		GroupField: Int32(6),
	},
	RepBytes: [][]byte{[]byte("sham"), []byte("wow")},
}

func init() {
	ext := &pb.Ext{
		Data: String("extension"),
	}
	if err := SetExtension(cloneTestMessage, pb.E_Ext_More, ext); err != nil {
		panic("SetExtension: " + err.Error())
	}
}

func TestClone(t *testing.T) {
	m := Clone(cloneTestMessage).(*pb.MyMessage)
	if !Equal(m, cloneTestMessage) {
		t.Fatalf("Clone(%v) = %v", cloneTestMessage, m)
	}

	// Verify it was a deep copy.
	*m.Inner.Port++
	if Equal(m, cloneTestMessage) {
		t.Error("Mutating clone changed the original")
	}
	// Byte fields and repeated fields should be copied.
	if &m.Pet[0] == &cloneTestMessage.Pet[0] {
		t.Error("Pet: repeated field not copied")
	}
	if &m.Others[0] == &cloneTestMessage.Others[0] {
		t.Error("Others: repeated field not copied")
	}
	if &m.Others[0].Value[0] == &cloneTestMessage.Others[0].Value[0] {
		t.Error("Others[0].Value: bytes field not copied")
	}
	if &m.RepBytes[0] == &cloneTestMessage.RepBytes[0] {
		t.Error("RepBytes: repeated field not copied")
	}
	if &m.RepBytes[0][0] == &cloneTestMessage.RepBytes[0][0] {
		t.Error("RepBytes[0]: bytes field not copied")
	}
}

func TestCloneNil(t *testing.T) {
	var m *pb.MyMessage
	if c := Clone(m); !Equal(m, c) {
		t.Errorf("Clone(%v) = %v", m, c)
	}
}

var mergeTests = []struct {
	src, dst, want Message
}{
	{
		src: &pb.MyMessage{
			Count: Int32(42),
		},
		dst: &pb.MyMessage{
			Name: String("Dave"),
		},
		want: &pb.MyMessage{
			Count: Int32(42),
			Name:  String("Dave"),
		},
	},
	{
		src: &pb.MyMessage{
			Inner: &pb.InnerMessage{
				Host:      String("hey"),
				Connected: Bool(true),
			},
			Pet: []string{"horsey"},
			Others: []*pb.OtherMessage{
				{
					Value: []byte("some bytes"),
				},
			},
		},
		dst: &pb.MyMessage{
			Inner: &pb.InnerMessage{
				Host: String("niles"),
				Port: Int32(9099),
			},
			Pet: []string{"bunny", "kitty"},
			Others: []*pb.OtherMessage{
				{
					Key: Int64(31415926535),
				},
				{
					// Explicitly test a src=nil field
					Inner: nil,
				},
			},
		},
		want: &pb.MyMessage{
			Inner: &pb.InnerMessage{
				Host:      String("hey"),
				Connected: Bool(true),
				Port:      Int32(9099),
			},
			Pet: []string{"bunny", "kitty", "horsey"},
			Others: []*pb.OtherMessage{
				{
					Key: Int64(31415926535),
				},
				{},
				{
					Value: []byte("some bytes"),
				},
			},
		},
	},
	{
		src: &pb.MyMessage{
			RepBytes: [][]byte{[]byte("wow")},
		},
		dst: &pb.MyMessage{
			Somegroup: &pb.MyMessage_SomeGroup{
				GroupField: Int32(6),
			},
			RepBytes: [][]byte{[]byte("sham")},
		},
		want: &pb.MyMessage{
			Somegroup: &pb.MyMessage_SomeGroup{
				GroupField: Int32(6),
			},
			RepBytes: [][]byte{[]byte("sham"), []byte("wow")},
		},
	},
	// Check that a scalar bytes field replaces rather than appends.
	{
		src:  &pb.OtherMessage{Value: []byte("foo")},
		dst:  &pb.OtherMessage{Value: []byte("bar")},
		want: &pb.OtherMessage{Value: []byte("foo")},
	},
	{
		src: &pb.MessageWithMap{
			NameMapping: map[int32]string{6: "Nigel"},
			MsgMapping: map[int64]*pb.FloatingPoint{
				0x4001: &pb.FloatingPoint{F: Float64(2.0)},
				0x4002: &pb.FloatingPoint{
					F: Float64(2.0),
				},
			},
			ByteMapping: map[bool][]byte{true: []byte("wowsa")},
		},
		dst: &pb.MessageWithMap{
			NameMapping: map[int32]string{
				6: "Bruce", // should be overwritten
				7: "Andrew",
			},
			MsgMapping: map[int64]*pb.FloatingPoint{
				0x4002: &pb.FloatingPoint{
					F:     Float64(3.0),
					Exact: Bool(true),
				}, // the entire message should be overwritten
			},
		},
		want: &pb.MessageWithMap{
			NameMapping: map[int32]string{
				6: "Nigel",
				7: "Andrew",
			},
			MsgMapping: map[int64]*pb.FloatingPoint{
				0x4001: &pb.FloatingPoint{F: Float64(2.0)},
				0x4002: &pb.FloatingPoint{
					F: Float64(2.0),
				},
			},
			ByteMapping: map[bool][]byte{true: []byte("wowsa")},
		},
	},
	// proto3 shouldn't merge zero values,
	// in the same way that proto2 shouldn't merge nils.
	{
		src: &proto3pb.Message{
			Name: "Aaron",
			Data: []byte(""), // zero value, but not nil
		},
		dst: &proto3pb.Message{
			HeightInCm: 176,
			Data:       []byte("texas!"),
		},
		want: &proto3pb.Message{
			Name:       "Aaron",
			HeightInCm: 176,
			Data:       []byte("texas!"),
		},
	},
	{ // Oneof fields should merge by assignment.
		src:  &pb.Communique{Union: &pb.Communique_Number{41}},
		dst:  &pb.Communique{Union: &pb.Communique_Name{"Bobby Tables"}},
		want: &pb.Communique{Union: &pb.Communique_Number{41}},
	},
	{ // Oneof nil is the same as not set.
		src:  &pb.Communique{},
		dst:  &pb.Communique{Union: &pb.Communique_Name{"Bobby Tables"}},
		want: &pb.Communique{Union: &pb.Communique_Name{"Bobby Tables"}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Number{1337}},
		dst:  &pb.Communique{},
		want: &pb.Communique{Union: &pb.Communique_Number{1337}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Col{pb.MyMessage_RED}},
		dst:  &pb.Communique{},
		want: &pb.Communique{Union: &pb.Communique_Col{pb.MyMessage_RED}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Data{[]byte("hello")}},
		dst:  &pb.Communique{},
		want: &pb.Communique{Union: &pb.Communique_Data{[]byte("hello")}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Msg{&pb.Strings{BytesField: []byte{1, 2, 3}}}},
		dst:  &pb.Communique{},
		want: &pb.Communique{Union: &pb.Communique_Msg{&pb.Strings{BytesField: []byte{1, 2, 3}}}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Msg{}},
		dst:  &pb.Communique{},
		want: &pb.Communique{Union: &pb.Communique_Msg{}},
	},
	{
		src:  &pb.Communique{Union: &pb.Communique_Msg{&pb.Strings{StringField: String("123")}}},
		dst:  &pb.Communique{Union: &pb.Communique_Msg{&pb.Strings{BytesField: []byte{1, 2, 3}}}},
		want: &pb.Communique{Union: &pb.Communique_Msg{&pb.Strings{StringField: String("123"), BytesField: []byte{1, 2, 3}}}},
	},
	{
		src: &proto3pb.Message{
			Terrain: map[string]*proto3pb.Nested{
				"kay_a": &proto3pb.Nested{Cute: true},      // replace
				"kay_b": &proto3pb.Nested{Bunny: "rabbit"}, // insert
			},
		},
		dst: &proto3pb.Message{
			Terrain: map[string]*proto3pb.Nested{
				"kay_a": &proto3pb.Nested{Bunny: "lost"},  // replaced
				"kay_c": &proto3pb.Nested{Bunny: "bunny"}, // keep
			},
		},
		want: &proto3pb.Message{
			Terrain: map[string]*proto3pb.Nested{
				"kay_a": &proto3pb.Nested{Cute: true},
				"kay_b": &proto3pb.Nested{Bunny: "rabbit"},
				"kay_c": &proto3pb.Nested{Bunny: "bunny"},
			},
		},
	},
	{
		src: &pb.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
		dst: &pb.GoTest{},
		want: &pb.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
	},
	{
		src: &pb.GoTest{},
		dst: &pb.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
		want: &pb.GoTest{
			F_BoolRepeated:   []bool{},
			F_Int32Repeated:  []int32{},
			F_Int64Repeated:  []int64{},
			F_Uint32Repeated: []uint32{},
			F_Uint64Repeated: []uint64{},
			F_FloatRepeated:  []float32{},
			F_DoubleRepeated: []float64{},
			F_StringRepeated: []string{},
			F_BytesRepeated:  [][]byte{},
		},
	},
	{
		src: &pb.GoTest{
			F_BytesRepeated: [][]byte{nil, []byte{}, []byte{0}},
		},
		dst: &pb.GoTest{},
		want: &pb.GoTest{
			F_BytesRepeated: [][]byte{nil, []byte{}, []byte{0}},
		},
	},
	{
		src: &pb.MyMessage{
			Others: []*pb.OtherMessage{},
		},
		dst: &pb.MyMessage{},
		want: &pb.MyMessage{
			Others: []*pb.OtherMessage{},
		},
	},
}

func TestMerge(t *testing.T) {
	for _, m := range mergeTests {
		got := Clone(m.dst)
		if !Equal(got, m.dst) {
			t.Errorf("Clone()\ngot  %v\nwant %v", got, m.dst)
			continue
		}
		Merge(got, m.src)
		if !Equal(got, m.want) {
			t.Errorf("Merge(%v, %v)\ngot  %v\nwant %v", m.dst, m.src, got, m.want)
		}
	}
}
