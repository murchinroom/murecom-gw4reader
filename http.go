package main

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// MureaderMurecomHandler handles murecom request from mureader:
//
// Request:
//
//	POST /murecom
//
//	{
//		"prevPages": ["昨天很伤心", "昨天很难过],
//		"currentPages": ["今天很开心"],
//		"nextPages": ["明天很迷茫", "明天很无助", "明天很绝望"]
//	}
//
// Response:
//
//	{
//		"music": {
//			"Title": "昨天很难过",
//			"Artist": "张三",
//			"CoverImage": "http://example.com/cover.jpg",
//			"SourceUrl": "http://example.com/source.mp3"
//	}
func MureaderMurecomHandler(c *gin.Context) {
	var req MureaderMurecomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := murecom4reader(&req)
	if err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

func murecom4reader(req *MureaderMurecomRequest) (*MureaderMurecomResponse, error) {
	logger := slog.With("func", "murecom4reader").With("req", req.LogValue())

	if req == nil {
		return nil, errors.New("request is nil")
	}

	emotions := []Emotion{}

	for _, pages := range [][]string{req.PrevPages, req.CurrentPages, req.NextPages} {
		for _, page := range pages {
			emotion, err := emotext(page)
			if err != nil {
				logger.Error("murecom4reader: emotext error. Try to continue.", "err", err)
				continue
			}
			emotions = append(emotions, emotion)
		}
	}

	avgEmotion := Emotion{}

	if len(emotions) == 0 {
		emotions = append(emotions, Emotion{Valence: 0.5, Arousal: 0.5})
	} // assert: len(emotions) > 0

	for _, emotion := range emotions {
		avgEmotion.Valence += emotion.Valence
		avgEmotion.Arousal += emotion.Arousal
	}
	avgEmotion.Valence /= float64(len(emotions))
	avgEmotion.Arousal /= float64(len(emotions))

	tracks, err := murecom(avgEmotion)
	if err != nil {
		logger.Warn("murecom4reader: murecom error.", "err", err)
		return nil, fmt.Errorf("murecom4reader: murecom error: %w", err)
	}

	if len(tracks) == 0 {
		logger.Warn("murecom4reader: murecom returns no track.")
		return nil, fmt.Errorf("murecom4reader: murecom returns no track")
	} // assert: len(tracks) > 0

	var resp MureaderMurecomResponse

	resp.Music = &Music{
		Title:        tracks[0].Name,
		Artist:       tracks[0].Artist,
		CoverImage:   tracks[0].CoverImageURL,
		SourceUrl:    tracks[0].AudioFileURL,
		TrackEmotion: tracks[0].Emotion,
	}
	resp.TextEmotion = avgEmotion

	logger.Info("murecom4reader: success.", "resp", resp)

	return &resp, nil
}
