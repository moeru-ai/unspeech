package volcengine

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Volcengine V3 bidirectional TTS binary protocol.
// https://www.volcengine.com/docs/6561/1329505

type v3Event int32

const (
	v3EventStartConnection    v3Event = 1
	v3EventFinishConnection   v3Event = 2
	v3EventConnectionStarted  v3Event = 50
	v3EventConnectionFailed   v3Event = 51
	v3EventConnectionFinished v3Event = 52

	v3EventStartSession    v3Event = 100
	v3EventCancelSession   v3Event = 101
	v3EventFinishSession   v3Event = 102
	v3EventSessionStarted  v3Event = 150
	v3EventSessionCanceled v3Event = 151
	v3EventSessionFinished v3Event = 152
	v3EventSessionFailed   v3Event = 153

	v3EventTaskRequest      v3Event = 200
	v3EventTTSSentenceStart v3Event = 350
	v3EventTTSSentenceEnd   v3Event = 351
	v3EventTTSResponse      v3Event = 352
)

const (
	v3MsgTypeFullClient = 0b0001
	v3MsgTypeError      = 0b1111

	v3FlagWithEvent = 0b0100

	v3SerialJSON = 0b0001

	// v3HeaderUnit is the multiplier used by the protocol's header-size field
	// (bits 0–3 of byte 0 encode header size in 4-byte units).
	v3HeaderUnit = 4
	// v3HalfByteMask masks the lower 4 bits of a header byte.
	v3HalfByteMask = 0x0f
	// v3MinHeaderSize is the protocol's minimum 4-byte header length.
	v3MinHeaderSize = 4
)

type v3Frame struct {
	msgType   byte
	flags     byte
	serial    byte
	compress  byte
	event     v3Event
	sessionID string
	errorCode uint32
	payload   []byte
}

// hasSessionID reports whether the event carries a session_id field on the
// wire. Connection-level events do not; session and data events do.
func hasSessionID(e v3Event) bool {
	switch e { //nolint:exhaustive
	case v3EventStartSession, v3EventCancelSession, v3EventFinishSession,
		v3EventSessionStarted, v3EventSessionCanceled, v3EventSessionFinished, v3EventSessionFailed,
		v3EventTaskRequest, v3EventTTSSentenceStart, v3EventTTSSentenceEnd, v3EventTTSResponse:
		return true
	}

	return false
}

// hasConnectionID reports whether the event carries a connection_id field on
// the wire. Only ConnectionStarted/Failed/Finished do.
func hasConnectionID(e v3Event) bool {
	switch e { //nolint:exhaustive
	case v3EventConnectionStarted, v3EventConnectionFailed, v3EventConnectionFinished:
		return true
	}

	return false
}

func encodeV3Frame(f v3Frame) []byte {
	// 4-byte header: version(1)<<4 | header_size(1), msgType<<4 | flags,
	// serial<<4 | compress, reserved.
	header := []byte{
		(1 << v3HeaderUnit) | 1,
		(f.msgType << v3HeaderUnit) | (f.flags & v3HalfByteMask),
		(f.serial << v3HeaderUnit) | (f.compress & v3HalfByteMask),
		0x00,
	}

	buf := make([]byte, 0, len(header)+8+len(f.sessionID)+len(f.payload))
	buf = append(buf, header...)

	if f.flags&v3FlagWithEvent != 0 {
		buf = binary.BigEndian.AppendUint32(buf, uint32(f.event))
	}

	if hasSessionID(f.event) {
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(f.sessionID)))
		buf = append(buf, f.sessionID...)
	}

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(f.payload)))
	buf = append(buf, f.payload...)

	return buf
}

func decodeV3Frame(data []byte) (v3Frame, error) {
	var f v3Frame
	if len(data) < v3MinHeaderSize {
		return f, errors.New("v3: frame too short")
	}

	f.msgType = data[1] >> v3HeaderUnit
	f.flags = data[1] & v3HalfByteMask
	f.serial = data[2] >> v3HeaderUnit
	f.compress = data[2] & v3HalfByteMask

	headerSize := max(int(data[0]&v3HalfByteMask)*v3HeaderUnit, v3MinHeaderSize)

	p := headerSize

	if f.msgType == v3MsgTypeError {
		if len(data) < p+8 {
			return f, errors.New("v3: error frame truncated")
		}

		f.errorCode = binary.BigEndian.Uint32(data[p : p+4])
		p += 4

		payloadSize := int(binary.BigEndian.Uint32(data[p : p+4]))
		p += 4

		if len(data) < p+payloadSize {
			return f, errors.New("v3: error payload truncated")
		}

		f.payload = data[p : p+payloadSize]

		return f, nil
	}

	if f.flags&v3FlagWithEvent != 0 {
		if len(data) < p+4 {
			return f, errors.New("v3: event field truncated")
		}

		f.event = v3Event(int32(binary.BigEndian.Uint32(data[p : p+4])))
		p += 4
	}

	switch {
	case hasSessionID(f.event):
		if len(data) < p+4 {
			return f, errors.New("v3: session id size truncated")
		}

		size := int(binary.BigEndian.Uint32(data[p : p+4]))
		p += 4

		if len(data) < p+size {
			return f, errors.New("v3: session id truncated")
		}

		f.sessionID = string(data[p : p+size])
		p += size
	case hasConnectionID(f.event):
		if len(data) < p+4 {
			return f, errors.New("v3: connection id size truncated")
		}

		size := int(binary.BigEndian.Uint32(data[p : p+4]))
		p += 4

		if len(data) < p+size {
			return f, errors.New("v3: connection id truncated")
		}

		// connection id is read but not retained — we don't need it after handshake.
		p += size
	}

	if len(data) < p+4 {
		return f, fmt.Errorf("v3: payload size truncated at event %d", f.event)
	}

	payloadSize := int(binary.BigEndian.Uint32(data[p : p+4]))
	p += 4

	if len(data) < p+payloadSize {
		return f, fmt.Errorf("v3: payload truncated at event %d (want %d, got %d)", f.event, payloadSize, len(data)-p)
	}

	f.payload = data[p : p+payloadSize]

	return f, nil
}
