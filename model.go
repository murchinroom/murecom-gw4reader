package main

import (
	"fmt"
	"murecom-gw4reader/ellipsis"
	"strings"

	"golang.org/x/exp/slog"
)

// MureaderMurecomRequest is the request from mureader
//
// from https://github.com/murchinroom/mureaderui/blob/main/lib/model.dart
type MureaderMurecomRequest struct {
	PrevPages    []string
	CurrentPages []string
	NextPages    []string
}

func (r MureaderMurecomRequest) LogValue() slog.Value {
	var sb strings.Builder

	sb.WriteString("{ PrevPg: [")
	for _, p := range r.PrevPages {
		sb.WriteString(ellipsis.Ending(p, 7))
		sb.WriteString(", ")
	}
	sb.WriteString("], CurrPg: [")
	for _, p := range r.CurrentPages {
		sb.WriteString(ellipsis.Centering(p, 7))
		sb.WriteString(", ")
	}
	sb.WriteString("], NextPg: [")
	for _, p := range r.NextPages {
		sb.WriteString(ellipsis.Starting(p, 7))
		sb.WriteString(", ")
	}
	sb.WriteString("] }")
	return slog.StringValue(sb.String())
}

// MureaderMurecomResponse is the response to mureader
//
// from https://github.com/murchinroom/mureaderui/blob/main/lib/model.dart
type MureaderMurecomResponse struct {
	Music       *Music
	TextEmotion Emotion // 分析的到的输入文本的情绪
}

func (r MureaderMurecomResponse) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("{ Music: %s, TextEmotion: %s }", r.Music.Title, r.TextEmotion.LogValue()))
}

// Music is the track for mureader
//
// from https://github.com/murchinroom/mureaderui/blob/main/lib/model.dart
type Music struct {
	Title      string
	Artist     string
	CoverImage string
	SourceUrl  string

	TrackEmotion Emotion
}

// MusicstoreTrack is the track from musicstore
//
// from https://github.com/murchinroom/musicstore/blob/main/model/model.go
type MusicstoreTrack struct {
	// orm.BasicModel

	Name          string
	Artist        string
	Album         string
	CoverImageURL string
	AudioFileURL  string

	Emotion Emotion `gorm:"embedded"`
}

// Emotion is the emotion of the MusicstoreTrack
//
// from https://github.com/murchinroom/mureaderui/blob/main/lib/model.dart
type Emotion struct {
	Valence float64 `json:"valence"`
	Arousal float64 `json:"arousal"`
}

func (e Emotion) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("{ Valence: %.2f, Arousal: %.2f }", e.Valence, e.Arousal))
}
