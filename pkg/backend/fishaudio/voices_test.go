package fishaudio

import (
	"testing"
)

func TestMapVoice(t *testing.T) {
	t.Parallel()

	voice := mapVoice(voiceModel{
		ID:          "802e3bc2b27e49c2995d23ef70e6ac89",
		Title:       "Example Voice",
		Description: "An example community voice model",
		Languages:   []string{"en", "zh", "xx"},
		Tags:        []string{"anime"},
		Samples: []voiceModelSample{
			{Title: "sample", Audio: "https://example.com/sample.mp3"},
		},
		Author:    voiceModelAuthor{ID: "author-id", Nickname: "Author"},
		LikeCount: 42,
		TaskCount: 1000,
	})

	if voice.ID != "802e3bc2b27e49c2995d23ef70e6ac89" {
		t.Fatalf("ID = %q", voice.ID)
	}
	if voice.Name != "Example Voice" {
		t.Fatalf("Name = %q", voice.Name)
	}
	if voice.PreviewAudioURL != "https://example.com/sample.mp3" {
		t.Fatalf("PreviewAudioURL = %q", voice.PreviewAudioURL)
	}
	if len(voice.Languages) != 3 {
		t.Fatalf("Languages = %+v, want 3 entries", voice.Languages)
	}
	if voice.Languages[0].Code != "en" || voice.Languages[0].Title != "English" {
		t.Fatalf("Languages[0] = %+v", voice.Languages[0])
	}
	// Unknown codes fall back to the raw code as title.
	if voice.Languages[2].Code != "xx" || voice.Languages[2].Title != "xx" {
		t.Fatalf("Languages[2] = %+v", voice.Languages[2])
	}
	if voice.Labels["author"] != "Author" {
		t.Fatalf("Labels[author] = %v", voice.Labels["author"])
	}
	if len(voice.CompatibleModels) != 4 {
		t.Fatalf("CompatibleModels = %+v", voice.CompatibleModels)
	}
}

func TestMapVoiceWithoutSamples(t *testing.T) {
	t.Parallel()

	voice := mapVoice(voiceModel{
		ID:    "id",
		Title: "No Samples",
	})

	if voice.PreviewAudioURL != "" {
		t.Fatalf("PreviewAudioURL = %q, want empty", voice.PreviewAudioURL)
	}
	if len(voice.Languages) != 0 {
		t.Fatalf("Languages = %+v, want empty", voice.Languages)
	}
}
