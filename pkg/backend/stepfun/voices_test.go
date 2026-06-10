package stepfun

import (
	"context"
	"testing"
)

func TestListVoicesIncludesStepfunSystemVoice(t *testing.T) {
	t.Parallel()

	voices, err := ListVoices(context.Background())
	if err != nil {
		t.Fatalf("ListVoices returned error: %v", err)
	}

	for _, voice := range voices {
		if voice.ID != "cixingnansheng" {
			continue
		}

		if voice.Name != "磁性男声" {
			t.Fatalf("Name = %q", voice.Name)
		}
		if len(voice.Formats) == 0 {
			t.Fatal("Formats is empty")
		}
		if len(voice.CompatibleModels) != 3 {
			t.Fatalf("CompatibleModels = %v", voice.CompatibleModels)
		}

		return
	}

	t.Fatal("cixingnansheng voice not found")
}
