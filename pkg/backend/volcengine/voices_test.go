package volcengine

import (
	"context"
	"testing"
)

// Floor counts from the 2026-05-17 ListSpeakers import (seed-tts-1.0: 345
// raw, 328 after dedupe by voice_type; seed-tts-2.0: 103). Source of truth
// is the Volcengine console `ListSpeakers` API — rebuild voices.json via
// scripts/import-volcengine-voices.sh when the catalogue rotates.
func TestListVoicesSeedTts10FloorAndIncompatibleExclusion(t *testing.T) {
	voices, err := ListVoices(context.Background(), "seed-tts-1.0")
	if err != nil {
		t.Fatalf("ListVoices: %v", err)
	}

	if len(voices) < 320 {
		t.Fatalf("seed-tts-1.0 voice count regressed below the 2026-05-17 import floor: got %d, want >= 320", len(voices))
	}

	streamingIncompatible := map[string]bool{
		"ICL_zh_female_bingruoshaonv_tob": true,
		"ICL_zh_female_huoponvhai_tob":    true,
		"ICL_zh_female_heainainai_tob":    true,
		"ICL_zh_female_linjuayi_tob":      true,
	}
	for _, v := range voices {
		if streamingIncompatible[v.ID] {
			t.Fatalf("voice %q is documented streaming-incompatible (https://www.volcengine.com/docs/6561/1257544) but leaked into the filtered result", v.ID)
		}
	}
}

func TestListVoicesSeedTts20FloorAndExclusivity(t *testing.T) {
	voices, err := ListVoices(context.Background(), "seed-tts-2.0")
	if err != nil {
		t.Fatalf("ListVoices: %v", err)
	}

	if len(voices) < 100 {
		t.Fatalf("seed-tts-2.0 voice count regressed below the 2026-05-17 import floor: got %d, want >= 100", len(voices))
	}

	for _, v := range voices {
		if !containsString(v.CompatibleModels, "seed-tts-2.0") {
			t.Fatalf("voice %q matched seed-tts-2.0 filter but its compatible_models %v does not include it", v.ID, v.CompatibleModels)
		}
	}
}

func TestListVoicesNoFilterReturnsUnionAndExcludesIncompatible(t *testing.T) {
	all, err := ListVoices(context.Background(), "")
	if err != nil {
		t.Fatalf("ListVoices: %v", err)
	}
	v10, _ := ListVoices(context.Background(), "seed-tts-1.0")
	v20, _ := ListVoices(context.Background(), "seed-tts-2.0")

	// Empty filter must return at least as many as the largest single-model
	// filter (any voice satisfying a filter also passes no-filter).
	if len(all) < len(v10) || len(all) < len(v20) {
		t.Fatalf("no-filter count %d must be >= per-model counts (1.0=%d, 2.0=%d)", len(all), len(v10), len(v20))
	}

	for _, v := range all {
		if v.ID == "ICL_zh_female_bingruoshaonv_tob" {
			t.Fatalf("supports_streaming=false voice leaked into no-filter result")
		}
	}
}

func TestListVoicesUnknownModelFilterReturnsEmpty(t *testing.T) {
	voices, err := ListVoices(context.Background(), "seed-tts-99.9-does-not-exist")
	if err != nil {
		t.Fatalf("ListVoices: %v", err)
	}
	if len(voices) != 0 {
		t.Fatalf("unknown model filter must return zero voices, got %d", len(voices))
	}
}
