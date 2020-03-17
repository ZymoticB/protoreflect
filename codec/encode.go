package codec

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// EncodeVarint writes a varint-encoded integer to the Buffer.
// This is the format for the
// int32, int64, uint32, uint64, bool, and enum
// protocol buffer types.
func (cb *Buffer) EncodeVarint(x uint64) error {
	for x >= 1<<7 {
		cb.buf = append(cb.buf, uint8(x&0x7f|0x80))
		x >>= 7
	}
	cb.buf = append(cb.buf, uint8(x))
	return nil
}

// EncodeTagAndWireType encodes the given field tag and wire type to the
// buffer. This combines the two values and then writes them as a varint.
func (cb *Buffer) EncodeTagAndWireType(tag protoreflect.FieldNumber, wireType int8) error {
	v := uint64((int64(tag) << 3) | int64(wireType))
	return cb.EncodeVarint(v)
}

// EncodeFixed64 writes a 64-bit integer to the Buffer.
// This is the format for the
// fixed64, sfixed64, and double protocol buffer types.
func (cb *Buffer) EncodeFixed64(x uint64) error {
	cb.buf = append(cb.buf,
		uint8(x),
		uint8(x>>8),
		uint8(x>>16),
		uint8(x>>24),
		uint8(x>>32),
		uint8(x>>40),
		uint8(x>>48),
		uint8(x>>56))
	return nil
}

// EncodeFixed32 writes a 32-bit integer to the Buffer.
// This is the format for the
// fixed32, sfixed32, and float protocol buffer types.
func (cb *Buffer) EncodeFixed32(x uint64) error {
	cb.buf = append(cb.buf,
		uint8(x),
		uint8(x>>8),
		uint8(x>>16),
		uint8(x>>24))
	return nil
}

// EncodeZigZag64 does zig-zag encoding to convert the given
// signed 64-bit integer into a form that can be expressed
// efficiently as a varint, even for negative values.
func EncodeZigZag64(v int64) uint64 {
	return (uint64(v) << 1) ^ uint64(v>>63)
}

// EncodeZigZag32 does zig-zag encoding to convert the given
// signed 32-bit integer into a form that can be expressed
// efficiently as a varint, even for negative values.
func EncodeZigZag32(v int32) uint64 {
	return uint64((uint32(v) << 1) ^ uint32((v >> 31)))
}

// EncodeRawBytes writes a count-delimited byte buffer to the Buffer.
// This is the format used for the bytes protocol buffer
// type and for embedded messages.
func (cb *Buffer) EncodeRawBytes(b []byte) error {
	if err := cb.EncodeVarint(uint64(len(b))); err != nil {
		return err
	}
	cb.buf = append(cb.buf, b...)
	return nil
}

// EncodeMessage writes the given message to the buffer.
func (cb *Buffer) EncodeMessage(pm proto.Message) error {
	bytes, err := marshalMessage(cb.buf, pm, cb.deterministic)
	if err != nil {
		return err
	}
	cb.buf = bytes
	return nil
}

// EncodeDelimitedMessage writes the given message to the buffer with a
// varint-encoded length prefix (the delimiter).
func (cb *Buffer) EncodeDelimitedMessage(pm proto.Message) error {
	bytes, err := marshalMessage(cb.tmp, pm, cb.deterministic)
	if err != nil {
		return err
	}
	// save truncated buffer if it was grown (so we can re-use it and
	// curtail future allocations)
	if cap(bytes) > cap(cb.tmp) {
		cb.tmp = bytes[:0]
	}
	return cb.EncodeRawBytes(bytes)
}

func marshalMessage(b []byte, pm proto.Message, deterministic bool) ([]byte, error) {
	opts := proto.MarshalOptions{
		Deterministic: deterministic,
	}
	return opts.MarshalAppend(b, pm)
}
