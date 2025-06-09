package ollama

import (
	"done-hub/common/requester"
	"done-hub/model"
	"done-hub/types"
	"encoding/json"
	"net/http"

	"done-hub/providers/base"
)

type OllamaProviderFactory struct{}

type OllamaProvider struct {
	base.BaseProvider
}

// 创建 OllamaProvider
func (f OllamaProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	config := getOllamaConfig()

	return &OllamaProvider{
		BaseProvider: base.BaseProvider{
			Config:    config,
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, RequestErrorHandle),
		},
	}
}

func getOllamaConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "",
		ChatCompletions: "/api/chat",
		Embeddings:      "/api/embeddings",
	}
}

// 请求错误处理
func RequestErrorHandle(resp *http.Response) *types.OpenAIError {
	errorResponse := &OllamaError{}
	err := json.NewDecoder(resp.Body).Decode(errorResponse)
	if err != nil {
		return nil
	}

	return errorHandle(errorResponse)
}

// 错误处理
func errorHandle(OllamaError *OllamaError) *types.OpenAIError {
	if OllamaError.Error == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: OllamaError.Error,
		Type:    "Ollama Error",
	}
}

// 获取请求头
func (p *OllamaProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	otherHeaders := p.Channel.Plugin.Data()["headers"]

	for key, value := range otherHeaders {
		headerValue, isString := value.(string)
		if !isString || headerValue == "" {
			continue
		}

		headers[key] = headerValue
	}

	return headers
}
