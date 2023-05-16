package main

import (
	"encoding/json"
	"murecom-gw4reader/ellipsis"
	"net/http"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

var EmotextServer = "http://127.0.0.1:9003"

func init() {
	if s, ok := os.LookupEnv("EMOTEXT_SERVER"); ok {
		EmotextServer = s
	}
}

// emotext requests emotext server.
func emotext(text string) (Emotion, error) {
	// curl -X POST -d '我今天很开心' http://localhost:9003
	logger := slog.With("func", "emotext").With("text", ellipsis.Centering(text, 11))

	body := strings.NewReader(text)

	req, err := http.NewRequest("POST", "http://localhost:9003", body)
	if err != nil {
		logger.Warn("emotext: http.NewRequest error.", "err", err)
		return Emotion{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Warn("emotext: http.DefaultClient.Do error.", "err", err)
		return Emotion{}, err
	}
	defer resp.Body.Close()

	// {"emotions": {"PA": 35.99, "NN": 8.13}, "polarity": {"neutrality": 2.01, "positive": 5.18, "negative": 2.71}, "va": {"valence": 0.65, "arousal": 0.67}}
	var emotextResp struct {
		VA Emotion `json:"va"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emotextResp); err != nil {
		logger.Warn("emotext: json.NewDecoder.Decode error.", "err", err)
		return Emotion{}, err
	}

	logger.Info("emotext: success.", "emotion", emotextResp.VA)

	return emotextResp.VA, nil
}
