package claude

import (
	"done-hub/common/requester"
	"done-hub/providers/base"
	"done-hub/types"
)

type ClaudeChatInterface interface {
	base.ProviderInterface
	CreateClaudeChat(request *ClaudeRequest) (*ClaudeResponse, *types.OpenAIErrorWithStatusCode)
	CreateClaudeChatStream(request *ClaudeRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode)
}
