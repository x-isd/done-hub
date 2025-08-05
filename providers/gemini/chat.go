package gemini

import (
	"done-hub/common"
	"done-hub/common/config"
	"done-hub/common/requester"
	"done-hub/common/utils"
	"done-hub/providers/base"
	"done-hub/types"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	GeminiVisionMaxImageNum = 16
)

type GeminiStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest

	key     string
	Context *gin.Context // 添加 Context 用于获取响应模型名称
}

type OpenAIStreamHandler struct {
	Usage     *types.Usage
	ModelName string
}

func (p *GeminiProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	if p.UseOpenaiAPI {
		return p.OpenAIProvider.CreateChatCompletion(request)
	}

	geminiRequest, errWithCode := ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	req, errWithCode := p.getChatRequest(geminiRequest, false)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	geminiChatResponse := &GeminiChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, geminiChatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return ConvertToChatOpenai(p, geminiChatResponse, request)
}

func (p *GeminiProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {

	channel := p.GetChannel()
	if p.UseOpenaiAPI {
		return p.OpenAIProvider.CreateChatCompletionStream(request)
	}

	geminiRequest, errWithCode := ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	req, errWithCode := p.getChatRequest(geminiRequest, false)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &GeminiStreamHandler{
		Usage:   p.Usage,
		Request: request,

		key:     channel.Key,
		Context: p.Context, // 传递 Context
	}

	return requester.RequestStream(p.Requester, resp, chatHandler.HandlerStream)
}

func (p *GeminiProvider) getChatRequest(geminiRequest *GeminiChatRequest, isRelay bool) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url := "generateContent"
	if geminiRequest.Stream {
		url = "streamGenerateContent?alt=sse"
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, geminiRequest.Model)

	// 获取请求头
	headers := p.GetRequestHeaders()
	if geminiRequest.Stream {
		headers["Accept"] = "text/event-stream"
	}

	var body any
	if isRelay {
		var exists bool
		rawData, exists := p.GetRawBody()
		if !exists {
			return nil, common.StringErrorWrapperLocal("request body not found", "request_body_not_found", http.StatusInternalServerError)
		}
		cleanedData, err := CleanGeminiRequestData(rawData, false)
		if err != nil {
			return nil, common.ErrorWrapper(err, "clean_relay_data_failed", http.StatusInternalServerError)
		}
		body = cleanedData
	} else {
		p.pluginHandle(geminiRequest)
		body = geminiRequest
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(body), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

// CleanGeminiRequestData 清理 Gemini 请求数据中的不兼容字段
func CleanGeminiRequestData(rawData []byte, isVertexAI bool) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	// 清理 contents 中的 function_call 和 function_response 字段中的 id
	if contents, ok := data["contents"].([]interface{}); ok {
		for _, content := range contents {
			if contentMap, ok := content.(map[string]interface{}); ok {
				if parts, ok := contentMap["parts"].([]interface{}); ok {
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							// 检查所有可能的字段名：functionCall, function_call
							fieldNames := []string{"functionCall", "function_call"}
							for _, fieldName := range fieldNames {
								if functionCall, ok := partMap[fieldName].(map[string]interface{}); ok {
									delete(functionCall, "id")
								}
							}

							// 检查所有可能的 function_response 字段名：functionResponse, function_response
							responseFieldNames := []string{"functionResponse", "function_response"}
							for _, fieldName := range responseFieldNames {
								if functionResponse, ok := partMap[fieldName].(map[string]interface{}); ok {
									delete(functionResponse, "id")
								}
							}
						}
					}
				}
			}
		}
	}

	// 如果是 Vertex AI，还需要清理 tools 中的 tool_type 字段
	if isVertexAI {
		if tools, ok := data["tools"].([]interface{}); ok {
			var validTools []interface{}
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					// 移除所有可能的 tool_type 相关字段，因为 Gemini API 不需要
					delete(toolMap, "tool_type")
					delete(toolMap, "toolType")
					delete(toolMap, "type")

					// 清理 functionDeclarations 中的 $schema 字段
					if functionDeclarations, ok := toolMap["functionDeclarations"].([]interface{}); ok {
						for _, funcDecl := range functionDeclarations {
							if funcDeclMap, ok := funcDecl.(map[string]interface{}); ok {
								if parameters, ok := funcDeclMap["parameters"].(map[string]interface{}); ok {
									// 移除 Vertex AI 不支持的 $schema 字段
									delete(parameters, "$schema")
									// 递归清理嵌套的 schema 对象
									cleanSchemaRecursively(parameters)
								}
							}
						}

						if len(functionDeclarations) == 0 {
							// 跳过空的工具
							continue
						}
					}

					// 检查工具是否有任何有效内容
					hasValidContent := false
					for key, value := range toolMap {
						if key == "functionDeclarations" {
							if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
								hasValidContent = true
								break
							}
						} else if value != nil {
							hasValidContent = true
							break
						}
					}

					if hasValidContent {
						validTools = append(validTools, toolMap)
					}
				}
			}

			// 如果没有有效工具，移除整个 tools 字段
			if len(validTools) == 0 {
				delete(data, "tools")
			} else {
				data["tools"] = validTools
			}
		}
	}

	return json.Marshal(data)
}

// cleanSchemaRecursively 递归清理 schema 对象中的 $schema 字段
func cleanSchemaRecursively(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		// 删除 $schema 字段
		delete(v, "$schema")

		// 递归处理所有值
		for _, value := range v {
			cleanSchemaRecursively(value)
		}
	case []interface{}:
		// 递归处理数组中的每个元素
		for _, item := range v {
			cleanSchemaRecursively(item)
		}
	}
}

func ConvertFromChatOpenai(request *types.ChatCompletionRequest) (*GeminiChatRequest, *types.OpenAIErrorWithStatusCode) {

	threshold := "BLOCK_NONE"

	// if strings.HasPrefix(request.Model, "gemini-2.0") && !strings.Contains(request.Model, "thinking") {
	// 	threshold = "OFF"
	// }

	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(request.Messages)),
		SafetySettings: []GeminiChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: threshold,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: threshold,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: threshold,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: threshold,
			},
			{
				Category:  "HARM_CATEGORY_CIVIC_INTEGRITY",
				Threshold: threshold,
			},
		},
		GenerationConfig: GeminiChatGenerationConfig{
			Temperature:        request.Temperature,
			TopP:               request.TopP,
			MaxOutputTokens:    request.MaxTokens,
			ResponseModalities: request.Modalities,
		},
	}

	if strings.HasPrefix(request.Model, "gemini-2.0-flash-exp") {
		geminiRequest.GenerationConfig.ResponseModalities = []string{"Text", "Image"}
	}

	if strings.HasSuffix(request.Model, "-tts") {
		geminiRequest.GenerationConfig.ResponseModalities = []string{"AUDIO"}
	}

	if request.Reasoning != nil {
		geminiRequest.GenerationConfig.ThinkingConfig = &ThinkingConfig{
			ThinkingBudget: &request.Reasoning.MaxTokens,
		}
	}

	if config.GeminiSettingsInstance.GetOpenThink(request.Model) {
		if geminiRequest.GenerationConfig.ThinkingConfig == nil {
			geminiRequest.GenerationConfig.ThinkingConfig = &ThinkingConfig{}
		}
		geminiRequest.GenerationConfig.ThinkingConfig.IncludeThoughts = true
	}

	functions := request.GetFunctions()

	if functions != nil {
		var geminiChatTools GeminiChatTools
		googleSearch := false
		codeExecution := false
		urlContext := false
		for _, function := range functions {
			if function.Name == "googleSearch" {
				googleSearch = true
				continue
			}
			if function.Name == "codeExecution" {
				codeExecution = true
				continue
			}
			if function.Name == "urlContext" {
				urlContext = true
				continue
			}

			if params, ok := function.Parameters.(map[string]interface{}); ok {
				if properties, ok := params["properties"].(map[string]interface{}); ok && len(properties) == 0 {
					function.Parameters = nil
				}
			}

			geminiChatTools.FunctionDeclarations = append(geminiChatTools.FunctionDeclarations, *function)
		}

		if codeExecution && len(geminiRequest.Tools) == 0 {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTools{
				CodeExecution: &GeminiCodeExecution{},
			})
		}
		if urlContext && len(geminiRequest.Tools) == 0 {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTools{
				UrlContext: &GeminiCodeExecution{},
			})
		}

		if googleSearch {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTools{
				GoogleSearch: &GeminiCodeExecution{},
			})
		}

		if len(geminiRequest.Tools) == 0 && len(geminiChatTools.FunctionDeclarations) > 0 {
			geminiRequest.Tools = append(geminiRequest.Tools, geminiChatTools)
		}
	}

	geminiContent, systemContent, err := OpenAIToGeminiChatContent(request.Messages)
	if err != nil {
		return nil, err
	}

	if systemContent != "" {
		geminiRequest.SystemInstruction = &GeminiChatContent{
			Parts: []GeminiPart{
				{Text: systemContent},
			},
		}
	}

	geminiRequest.Contents = geminiContent
	geminiRequest.Stream = request.Stream
	geminiRequest.Model = request.Model

	if request.ResponseFormat != nil && (request.ResponseFormat.Type == "json_schema" || request.ResponseFormat.Type == "json_object") {
		geminiRequest.GenerationConfig.ResponseMimeType = "application/json"

		if request.ResponseFormat.JsonSchema != nil && request.ResponseFormat.JsonSchema.Schema != nil {
			cleanedSchema := removeAdditionalPropertiesWithDepth(request.ResponseFormat.JsonSchema.Schema, 0)
			geminiRequest.GenerationConfig.ResponseSchema = cleanedSchema
		}
	}

	return &geminiRequest, nil
}

func removeAdditionalPropertiesWithDepth(schema interface{}, depth int) interface{} {
	if depth >= 5 {
		return schema
	}

	v, ok := schema.(map[string]interface{})
	if !ok || len(v) == 0 {
		return schema
	}

	// 如果type不为object和array，则直接返回
	if typeVal, exists := v["type"]; !exists || (typeVal != "object" && typeVal != "array") {
		return schema
	}

	delete(v, "title")
	// 删除 $schema 字段，因为 Gemini API 不支持
	delete(v, "$schema")

	// 处理format字段的限制 - Gemini API只支持STRING类型的"enum"和"date-time"格式
	if formatVal, exists := v["format"]; exists {
		if formatStr, ok := formatVal.(string); ok {
			if typeVal, typeExists := v["type"]; typeExists && typeVal == "string" {
				// 只保留Gemini支持的format
				if formatStr != "enum" && formatStr != "date-time" {
					delete(v, "format")
				}
			}
		}
	}

	switch v["type"] {
	case "object":
		delete(v, "additionalProperties")
		// 处理 properties
		if properties, ok := v["properties"].(map[string]interface{}); ok {
			for key, value := range properties {
				properties[key] = removeAdditionalPropertiesWithDepth(value, depth+1)
			}
		}
		for _, field := range []string{"allOf", "anyOf", "oneOf"} {
			if nested, ok := v[field].([]interface{}); ok {
				for i, item := range nested {
					nested[i] = removeAdditionalPropertiesWithDepth(item, depth+1)
				}
			}
		}
	case "array":
		if items, ok := v["items"].(map[string]interface{}); ok {
			v["items"] = removeAdditionalPropertiesWithDepth(items, depth+1)
		}
	}

	return v
}

func ConvertToChatOpenai(provider base.ProviderInterface, response *GeminiChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	// 获取响应中应该使用的模型名称
	responseModel := provider.GetResponseModelName(request.Model)

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.ResponseId,
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Model:   responseModel,
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Candidates)),
	}

	if len(response.Candidates) == 0 {
		errWithCode = common.StringErrorWrapper("no candidates", "no_candidates", http.StatusInternalServerError)
		return
	}

	for _, candidate := range response.Candidates {
		openaiResponse.Choices = append(openaiResponse.Choices, candidate.ToOpenAIChoice(request))
	}

	usage := provider.GetUsage()
	*usage = ConvertOpenAIUsage(response.UsageMetadata)
	openaiResponse.Usage = usage

	return
}

// 转换为OpenAI聊天流式请求体
func (h *GeminiStreamHandler) HandlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var geminiResponse GeminiChatResponse
	err := json.Unmarshal(*rawLine, &geminiResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&geminiResponse.GeminiErrorResponse, h.key)
	if aiError != nil {
		errChan <- aiError
		return
	}

	h.convertToOpenaiStream(&geminiResponse, dataChan)

}

func (h *GeminiStreamHandler) convertToOpenaiStream(geminiResponse *GeminiChatResponse, dataChan chan string) {
	// 获取响应中应该使用的模型名称
	responseModel := h.Request.Model
	if h.Context != nil {
		responseModel = base.GetResponseModelNameFromContext(h.Context, h.Request.Model)
	}

	streamResponse := types.ChatCompletionStreamResponse{
		ID:      geminiResponse.ResponseId,
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   responseModel,
		// Choices: choices,
	}

	choices := make([]types.ChatCompletionStreamChoice, 0, len(geminiResponse.Candidates))

	isStop := false
	for _, candidate := range geminiResponse.Candidates {
		if candidate.FinishReason != nil && *candidate.FinishReason == "STOP" {
			isStop = true
			candidate.FinishReason = nil
		}
		choices = append(choices, candidate.ToOpenAIStreamChoice(h.Request))
	}

	if len(choices) > 0 && (choices[0].Delta.ToolCalls != nil || choices[0].Delta.FunctionCall != nil) {
		choices := choices[0].ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := streamResponse
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	} else {
		streamResponse.Choices = choices
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}

	if isStop {
		streamResponse.Choices = []types.ChatCompletionStreamChoice{
			{
				FinishReason: types.FinishReasonStop,
				Delta: types.ChatCompletionStreamChoiceDelta{
					Role: types.ChatMessageRoleAssistant,
				},
			},
		}
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}

	// 和ExecutableCode的tokens共用，所以跳过
	if geminiResponse.UsageMetadata == nil {
		return
	}

	h.Usage.PromptTokens = geminiResponse.UsageMetadata.PromptTokenCount

	// 计算 completion tokens，确保不为负数
	completionTokens := geminiResponse.UsageMetadata.CandidatesTokenCount + geminiResponse.UsageMetadata.ThoughtsTokenCount
	if completionTokens < 0 {
		completionTokens = 0
	}
	h.Usage.CompletionTokens = completionTokens
	h.Usage.CompletionTokensDetails.ReasoningTokens = geminiResponse.UsageMetadata.ThoughtsTokenCount

	// 如果 TotalTokenCount 为 0 但有 PromptTokenCount，则计算总数
	totalTokens := geminiResponse.UsageMetadata.TotalTokenCount
	if totalTokens == 0 && geminiResponse.UsageMetadata.PromptTokenCount > 0 {
		totalTokens = geminiResponse.UsageMetadata.PromptTokenCount + completionTokens
	}
	h.Usage.TotalTokens = totalTokens
}

const tokenThreshold = 1000000

var modelAdjustRatios = map[string]int{
	"gemini-1.5-pro":   2,
	"gemini-1.5-flash": 2,
}

// func adjustTokenCounts(modelName string, usage *GeminiUsageMetadata) {
// 	if usage.PromptTokenCount <= tokenThreshold && usage.CandidatesTokenCount <= tokenThreshold {
// 		return
// 	}

// 	currentRatio := 1
// 	for model, r := range modelAdjustRatios {
// 		if strings.HasPrefix(modelName, model) {
// 			currentRatio = r
// 			break
// 		}
// 	}

// 	if currentRatio == 1 {
// 		return
// 	}

// 	adjustTokenCount := func(count int) int {
// 		if count > tokenThreshold {
// 			return tokenThreshold + (count-tokenThreshold)*currentRatio
// 		}
// 		return count
// 	}

// 	if usage.PromptTokenCount > tokenThreshold {
// 		usage.PromptTokenCount = adjustTokenCount(usage.PromptTokenCount)
// 	}

// 	if usage.CandidatesTokenCount > tokenThreshold {
// 		usage.CandidatesTokenCount = adjustTokenCount(usage.CandidatesTokenCount)
// 	}

// 	usage.TotalTokenCount = usage.PromptTokenCount + usage.CandidatesTokenCount
// }

func ConvertOpenAIUsage(geminiUsage *GeminiUsageMetadata) types.Usage {
	if geminiUsage == nil {
		return types.Usage{}
	}

	// 计算 completion tokens，确保不为负数
	completionTokens := geminiUsage.CandidatesTokenCount + geminiUsage.ThoughtsTokenCount
	if completionTokens < 0 {
		completionTokens = 0
	}

	// 如果 TotalTokenCount 为 0 但有 PromptTokenCount，则计算总数
	totalTokens := geminiUsage.TotalTokenCount
	if totalTokens == 0 && geminiUsage.PromptTokenCount > 0 {
		totalTokens = geminiUsage.PromptTokenCount + completionTokens
	}

	return types.Usage{
		PromptTokens:     geminiUsage.PromptTokenCount,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,

		CompletionTokensDetails: types.CompletionTokensDetails{
			ReasoningTokens: geminiUsage.ThoughtsTokenCount,
		},
	}
}

func (p *GeminiProvider) pluginHandle(request *GeminiChatRequest) {
	if !p.UseCodeExecution {
		return
	}

	if len(request.Tools) > 0 {
		return
	}

	if p.Channel.Plugin == nil {
		return
	}

	request.Tools = append(request.Tools, GeminiChatTools{
		CodeExecution: &GeminiCodeExecution{},
	})

}
