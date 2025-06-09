package openai

import (
	"done-hub/common/config"
	"done-hub/common/requester"
	"done-hub/types"
	"net/http"
)

func (p *OpenAIProvider) CreateSpeech(request *types.SpeechAudioRequest) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.GetRequestTextBody(config.RelayModeAudioSpeech, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	var resp *http.Response
	resp, errWithCode = p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		return nil, requester.HandleErrorResp(resp, p.Requester.ErrorHandler, p.Requester.IsOpenAI)
	}

	p.Usage.TotalTokens = p.Usage.PromptTokens

	return resp, nil
}
