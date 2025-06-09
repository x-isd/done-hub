package minimax

import (
	"done-hub/common/requester"
	"done-hub/types"
	"encoding/json"
)

func (p *MiniMaxProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	conversionRequest(request)

	return p.OpenAIProvider.CreateChatCompletion(request)
}

func (p *MiniMaxProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	conversionRequest(request)

	return p.OpenAIProvider.CreateChatCompletionStream(request)
}

func conversionRequest(request *types.ChatCompletionRequest) {
	if len(request.Tools) > 0 {
		for _, tool := range request.Tools {
			if tool.Function.Parameters != nil {
				parameters, err := json.Marshal(tool.Function.Parameters)
				if err != nil {
					continue
				}
				tool.Function.Parameters = string(parameters)
			}
		}
	}
}
