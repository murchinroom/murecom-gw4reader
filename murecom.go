package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/exp/slog"
)

var MusicstoreMurecomServer = "http://127.0.0.1:8080/murecom"

func init() {
	if s, ok := os.LookupEnv("MUSICSTORE_MURECOM_SERVER"); ok {
		MusicstoreMurecomServer = s
	}
}

// murecom requests musicstore murecom server.
func murecom(emotion Emotion) ([]*MusicstoreTrack, error) {
	// curl 'localhost:8080/murecom?Valence=0.5&Arousal=0.5'

	logger := slog.With("func", "murecom").With("emotion", emotion)

	resp, err := http.Get(fmt.Sprintf("%s?Valence=%v&Arousal=%v",
		MusicstoreMurecomServer, emotion.Valence, emotion.Arousal))

	if err != nil {
		logger.Warn("murecom: http.Get error.", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	var respBody struct {
		Tracks []*MusicstoreTrack `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		logger.Warn("murecom: json.NewDecoder.Decode error.", "err", err)
		return nil, err
	}

	logger.Info("murecom: success.", "len(tracks)", len(respBody.Tracks))

	return respBody.Tracks, nil
}
