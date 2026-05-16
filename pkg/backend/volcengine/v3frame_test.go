package volcengine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV3FrameRoundTrip_StartConnection(t *testing.T) {
	in := v3Frame{
		msgType: v3MsgTypeFullClient,
		flags:   v3FlagWithEvent,
		serial:  v3SerialJSON,
		event:   v3EventStartConnection,
		payload: []byte(`{}`),
	}

	out, err := decodeV3Frame(encodeV3Frame(in))
	require.NoError(t, err)
	require.Equal(t, in.event, out.event)
	require.Equal(t, in.payload, out.payload)
	require.Empty(t, out.sessionID)
}

func TestV3FrameRoundTrip_TaskRequest(t *testing.T) {
	in := v3Frame{
		msgType:   v3MsgTypeFullClient,
		flags:     v3FlagWithEvent,
		serial:    v3SerialJSON,
		event:     v3EventTaskRequest,
		sessionID: "session-abc",
		payload:   []byte(`{"text":"hello"}`),
	}

	out, err := decodeV3Frame(encodeV3Frame(in))
	require.NoError(t, err)
	require.Equal(t, in.event, out.event)
	require.Equal(t, in.sessionID, out.sessionID)
	require.Equal(t, in.payload, out.payload)
}

func TestV3FrameRoundTrip_TTSResponseAudio(t *testing.T) {
	// Audio-only response from the server carries the session id and raw bytes.
	in := v3Frame{
		msgType:   v3MsgTypeFullClient,
		flags:     v3FlagWithEvent,
		serial:    v3SerialJSON,
		event:     v3EventTTSResponse,
		sessionID: "sid",
		payload:   []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	out, err := decodeV3Frame(encodeV3Frame(in))
	require.NoError(t, err)
	require.Equal(t, in.event, out.event)
	require.Equal(t, in.sessionID, out.sessionID)
	require.Equal(t, in.payload, out.payload)
}

func TestDecodeV3Frame_RejectsShortInput(t *testing.T) {
	_, err := decodeV3Frame([]byte{0x11})
	require.Error(t, err)
}

func TestDecodeV3Frame_TruncatedPayload(t *testing.T) {
	frame := encodeV3Frame(v3Frame{
		msgType:   v3MsgTypeFullClient,
		flags:     v3FlagWithEvent,
		serial:    v3SerialJSON,
		event:     v3EventTaskRequest,
		sessionID: "sid",
		payload:   []byte(`{"text":"hello"}`),
	})

	_, err := decodeV3Frame(frame[:len(frame)-3])
	require.Error(t, err)
}
