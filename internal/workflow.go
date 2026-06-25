package internal

import (
	"fmt"
	"time"

	"pi/pkg/logger"
)

func ExecuteWorkflow(input string, primaryModel string, chatHistory []Content) ([]Content, string, bool, error) {
	if cfg, err := LoadConfig(); err == nil {
		if cfg.GoogleAPIKey != "" {
			ApiKey = cfg.GoogleAPIKey
		}
		if cfg.Model1 != "" {
			primaryModel = cfg.Model1
		}
	}

	currentPrompt := input
	chatHistory = append(chatHistory, Content{
		Role:  "user",
		Parts: []Part{{Text: currentPrompt}},
	})

	for {
		logger.Log.Info("AI 코드 생성 중...\n")
		var aiResponse string
		var callErr error
		for attempt := 1; attempt <= 3; attempt++ {
			aiResponse, callErr = CallGemini(primaryModel, chatHistory)
			if callErr == nil {
				break
			}
			logger.Log.Warn("[API 오류 감지] %s 모델 호출 실패(시도 %d/3): %v. 재시도 진행...\n", primaryModel, attempt, callErr)
			if attempt < 3 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		if callErr != nil {
			return chatHistory, "", false, fmt.Errorf("API 최종 호출 오류: %w", callErr)
		}

		lang, code := ExtractCodeBlock(aiResponse)

		if lang != "" && code != "" {
			success, feedbackResponse, updatedHistory, err := ExecuteGenericWorkflow(lang, code, input, chatHistory, aiResponse, primaryModel)
			if err != nil {
				return chatHistory, "", false, fmt.Errorf("실행 워크플로 오류: %w", err)
			}
			if !success {
				chatHistory = updatedHistory
				continue
			}

			chatHistory = append(chatHistory, Content{
				Role:  "model",
				Parts: []Part{{Text: aiResponse}},
			})
			chatHistory = append(chatHistory, Content{
				Role:  "user",
				Parts: []Part{{Text: "성공적으로 실행됨. 다음은 실행 결과 분석 보고서:\n" + feedbackResponse}},
			})
			return chatHistory, feedbackResponse, true, nil
		} else {
			chatHistory = append(chatHistory, Content{
				Role:  "model",
				Parts: []Part{{Text: aiResponse}},
			})
			return chatHistory, aiResponse, false, nil
		}
	}
}
