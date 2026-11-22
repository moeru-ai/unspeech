package fishaudio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		Licensed:  true,
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
	if voice.Labels["licensed"] != true {
		t.Fatalf("Labels[licensed] = %v", voice.Labels["licensed"])
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

func TestListVoicesForwardsFiltersAndTotal(t *testing.T) {
	t.Parallel()

	var gotQuery url.Values
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		gotAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":1240,"has_more":true,"items":[{"_id":"voice-1","title":"Voice","languages":["zh"],"licensed":true}]}`))
	}))

	defer server.Close()

	query := url.Values{
		"language":    {"zh", "ja"},
		"tag":         {"female"},
		"licensed":    {"true"},
		"title":       {"calm"},
		"page_number": {"3"},
		"page_size":   {"20"},
		"unknown":     {"dropped"},
	}

	response, err := listVoices(context.Background(), server.Client(), server.URL, VoicesCredentials{APIKey: "test-key"}, query)
	if err != nil {
		t.Fatalf("listVoices() error = %v", err)
	}

	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if langs := gotQuery["language"]; len(langs) != 2 || langs[0] != "zh" || langs[1] != "ja" {
		t.Fatalf("language = %v", langs)
	}

	for key, want := range map[string]string{
		"tag": "female", "licensed": "true", "title": "calm", "page_number": "3", "page_size": "20", "sort_by": "task_count",
	} {
		if got := gotQuery.Get(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
	if gotQuery.Has("unknown") {
		t.Fatalf("unknown filter was forwarded: %v", gotQuery)
	}

	if response.Total == nil || *response.Total != 1240 {
		t.Fatalf("Total = %v", response.Total)
	}
	if response.HasMore == nil || !*response.HasMore {
		t.Fatalf("HasMore = %v", response.HasMore)
	}
	if len(response.Voices) != 1 || response.Voices[0].Labels["licensed"] != true {
		t.Fatalf("Voices = %+v", response.Voices)
	}
}
