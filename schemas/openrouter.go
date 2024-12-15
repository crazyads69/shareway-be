package schemas

// Define OpenRouter API response
type Choice struct {
	LogProbs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
	Index        int         `json:"index"`
	Message      struct {
		Role    string `json:"role"`
		Content string `json:"content"`
		Refusal string `json:"refusal"`
	} `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenRouterResponse struct {
	ID                string      `json:"id"`
	Provider          string      `json:"provider"`
	Model             string      `json:"model"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Choices           []Choice    `json:"choices"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
	Usage             Usage       `json:"usage"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type OpenRouterRequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
