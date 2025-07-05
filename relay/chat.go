package relay

import (
	"done-hub/common"
	"done-hub/common/config"
	"done-hub/common/requester"
	"done-hub/common/utils"
	providersBase "done-hub/providers/base"
	"done-hub/safty"
	"done-hub/types"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type relayChat struct {
	relayBase
	chatRequest types.ChatCompletionRequest
}

func NewRelayChat(c *gin.Context) *relayChat {
	relay := &relayChat{
		relayBase: relayBase{
			allowHeartbeat: true,
			c:              c,
		},
	}
	return relay
}

func (r *relayChat) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.chatRequest); err != nil {
		return err
	}

	if r.chatRequest.MaxTokens < 0 || r.chatRequest.MaxTokens > math.MaxInt32/2 {
		return errors.New("max_tokens is invalid")
	}

	if r.chatRequest.Tools != nil {
		r.c.Set("skip_only_chat", true)
	}

	if !r.chatRequest.Stream {
		r.chatRequest.StreamOptions = nil
	}

	// 过滤空content的消息，保持所有渠道行为一致
	r.filterEmptyContentMessages()

	r.setOriginalModel(r.chatRequest.Model)

	otherArg := r.getOtherArg()

	if otherArg == "search" {
		handleSearch(r.c, &r.chatRequest)
		return nil
	}

	return nil
}

func (r *relayChat) getRequest() interface{} {
	return &r.chatRequest
}

func (r *relayChat) IsStream() bool {
	return r.chatRequest.Stream
}

func (r *relayChat) getPromptTokens() (int, error) {
	channel := r.provider.GetChannel()
	return common.CountTokenMessages(r.chatRequest.Messages, r.modelName, channel.PreCost), nil
}

func (r *relayChat) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	chatProvider, ok := r.provider.(providersBase.ChatInterface)
	if !ok {
		err = common.StringErrorWrapperLocal("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.chatRequest.Model = r.modelName
	// 内容审查
	if config.EnableSafe {
		for _, message := range r.chatRequest.Messages {
			if message.Content != nil {
				CheckResult, _ := safty.CheckContent(message.Content)
				if !CheckResult.IsSafe {
					err = common.StringErrorWrapperLocal(CheckResult.Reason, CheckResult.Code, http.StatusBadRequest)
					done = true
					return
				}
			}
		}
	}

	if r.chatRequest.Stream {
		var response requester.StreamReaderInterface[string]
		response, err = chatProvider.CreateChatCompletionStream(&r.chatRequest)
		if err != nil {
			return
		}

		if r.heartbeat != nil {
			r.heartbeat.Stop()
		}

		doneStr := func() string {
			return r.getUsageResponse()
		}

		var firstResponseTime time.Time
		firstResponseTime, err = responseStreamClient(r.c, response, doneStr)
		r.SetFirstResponseTime(firstResponseTime)
	} else {
		var response *types.ChatCompletionResponse
		response, err = chatProvider.CreateChatCompletion(&r.chatRequest)
		if err != nil {
			return
		}

		if r.heartbeat != nil {
			r.heartbeat.Stop()
		}

		err = responseJsonClient(r.c, response)

	}

	if err != nil {
		done = true
	}

	return
}

// filterEmptyContentMessages 过滤掉content为空的消息，保持所有渠道行为一致
func (r *relayChat) filterEmptyContentMessages() {
	if len(r.chatRequest.Messages) == 0 {
		return
	}

	filteredMessages := make([]types.ChatCompletionMessage, 0, len(r.chatRequest.Messages))

	for _, message := range r.chatRequest.Messages {
		if r.isMessageContentEmpty(message) {
			continue // 跳过空content的消息
		}
		filteredMessages = append(filteredMessages, message)
	}

	r.chatRequest.Messages = filteredMessages
}

// isMessageContentEmpty 检查消息的content是否为空
// 只有在content完全为空且没有其他重要字段时才认为消息为空
func (r *relayChat) isMessageContentEmpty(message types.ChatCompletionMessage) bool {
	// 如果消息有tool_calls、function_call等重要字段，不认为是空消息
	if message.ToolCalls != nil || message.FunctionCall != nil || message.ToolCallID != "" {
		return false
	}

	// 如果content为nil，认为是空的
	if message.Content == nil {
		return true
	}

	// 如果是字符串类型的content
	if contentStr, ok := message.Content.(string); ok {
		// 去除空白字符后检查是否为空
		return strings.TrimSpace(contentStr) == ""
	}

	// 如果是数组类型的content（多模态消息）
	contentParts := message.ParseContent()
	if len(contentParts) == 0 {
		return true
	}

	// 检查所有content part是否都为空
	hasNonEmptyContent := false
	for _, part := range contentParts {
		if part.Type == types.ContentTypeText {
			if strings.TrimSpace(part.Text) != "" {
				hasNonEmptyContent = true
				break
			}
		} else {
			// 非文本类型的content（如图片、音频等）认为是有效的
			hasNonEmptyContent = true
			break
		}
	}

	return !hasNonEmptyContent
}

func (r *relayChat) getUsageResponse() string {
	if r.chatRequest.StreamOptions != nil && r.chatRequest.StreamOptions.IncludeUsage {
		usageResponse := types.ChatCompletionStreamResponse{
			ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
			Object:  "chat.completion.chunk",
			Created: utils.GetTimestamp(),
			Model:   r.chatRequest.Model,
			Choices: []types.ChatCompletionStreamChoice{},
			Usage:   r.provider.GetUsage(),
		}

		responseBody, err := json.Marshal(usageResponse)
		if err != nil {
			return ""
		}

		return string(responseBody)
	}

	return ""
}
