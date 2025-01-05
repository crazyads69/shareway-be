package schemas

// Request represents the main request structure
type OpenRouterRequestBody struct {
	Messages          []Message            `json:"messages,omitempty"`
	Prompt            string               `json:"prompt,omitempty"`
	Model             string               `json:"model,omitempty"`
	ResponseFormat    *ResponseFormat      `json:"response_format,omitempty"`
	Stop              interface{}          `json:"stop,omitempty"` // Can be string or []string
	Stream            *bool                `json:"stream,omitempty"`
	MaxTokens         *int                 `json:"max_tokens,omitempty"`
	Temperature       *float64             `json:"temperature,omitempty"`
	Tools             []Tool               `json:"tools,omitempty"`
	ToolChoice        interface{}          `json:"tool_choice,omitempty"` // Can be string or ToolChoice
	Seed              *int                 `json:"seed,omitempty"`
	TopP              *float64             `json:"top_p,omitempty"`
	TopK              *int                 `json:"top_k,omitempty"`
	FrequencyPenalty  *float64             `json:"frequency_penalty,omitempty"`
	PresencePenalty   *float64             `json:"presence_penalty,omitempty"`
	RepetitionPenalty *float64             `json:"repetition_penalty,omitempty"`
	LogitBias         map[int]int          `json:"logit_bias,omitempty"`
	TopLogprobs       *int                 `json:"top_logprobs,omitempty"`
	MinP              *float64             `json:"min_p,omitempty"`
	TopA              *float64             `json:"top_a,omitempty"`
	Prediction        *PredictionRequest   `json:"prediction,omitempty"`
	Transforms        []string             `json:"transforms,omitempty"`
	Models            []string             `json:"models,omitempty"`
	Route             string               `json:"route,omitempty"`
	Provider          *ProviderPreferences `json:"provider,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type PredictionRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"` // Can be string or []ContentPart
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type FunctionDescription struct {
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name"`
	Parameters  interface{} `json:"parameters"` // JSON Schema object
}

type Tool struct {
	Type     string              `json:"type"`
	Function FunctionDescription `json:"function"`
}

type ToolChoice struct {
	Type     string `json:"type"`
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}

// Response represents the main response structure
type OpenRouterResponse struct {
	ID                string   `json:"id"`
	Choices           []Choice `json:"choices"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	Usage             *Usage   `json:"usage,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	FinishReason string         `json:"finish_reason"`
	Index        int            `json:"index"`
	Message      *Message       `json:"message,omitempty"`
	Delta        *Delta         `json:"delta,omitempty"`
	Text         string         `json:"text,omitempty"`
	Error        *ErrorResponse `json:"error,omitempty"`
}

type Delta struct {
	Content   string     `json:"content,omitempty"`
	Role      string     `json:"role,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ErrorResponse struct {
	Code     int                    `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ProviderPreferences represents the preferences for selecting and configuring providers
type ProviderPreferences struct {
	AllowFallbacks    *bool    `json:"allow_fallbacks,omitempty"`
	RequireParameters *bool    `json:"require_parameters,omitempty"`
	DataCollection    *string  `json:"data_collection,omitempty"`
	Order             []string `json:"order,omitempty"`
	Ignore            []string `json:"ignore,omitempty"`
	Quantizations     []string `json:"quantizations,omitempty"`
}

// Constants for DataCollection
const (
	DataCollectionAllow = "allow"
	DataCollectionDeny  = "deny"
)

// Constants for Quantizations
const (
	QuantizationInt4    = "int4"
	QuantizationInt8    = "int8"
	QuantizationFP6     = "fp6"
	QuantizationFP8     = "fp8"
	QuantizationFP16    = "fp16"
	QuantizationBF16    = "bf16"
	QuantizationUnknown = "unknown"
)

// ProviderName represents the available provider names
type ProviderName string

// Constants for ProviderName
const (
	ProviderOpenAI         ProviderName = "OpenAI"
	ProviderAnthropic      ProviderName = "Anthropic"
	ProviderGoogle         ProviderName = "Google"
	ProviderGoogleAIStudio ProviderName = "Google AI Studio"
	ProviderAmazonBedrock  ProviderName = "Amazon Bedrock"
	ProviderGroq           ProviderName = "Groq"
	ProviderSambaNova      ProviderName = "SambaNova"
	ProviderCohere         ProviderName = "Cohere"
	ProviderMistral        ProviderName = "Mistral"
	ProviderTogether       ProviderName = "Together"
	ProviderTogether2      ProviderName = "Together 2"
	ProviderFireworks      ProviderName = "Fireworks"
	ProviderDeepInfra      ProviderName = "DeepInfra"
	ProviderLepton         ProviderName = "Lepton"
	ProviderNovita         ProviderName = "Novita"
	ProviderAvian          ProviderName = "Avian"
	ProviderLambda         ProviderName = "Lambda"
	ProviderAzure          ProviderName = "Azure"
	ProviderModal          ProviderName = "Modal"
	ProviderAnyScale       ProviderName = "AnyScale"
	ProviderReplicate      ProviderName = "Replicate"
	ProviderPerplexity     ProviderName = "Perplexity"
	ProviderRecursal       ProviderName = "Recursal"
	ProviderOctoAI         ProviderName = "OctoAI"
	ProviderDeepSeek       ProviderName = "DeepSeek"
	ProviderInfermatic     ProviderName = "Infermatic"
	ProviderAI21           ProviderName = "AI21"
	ProviderFeatherless    ProviderName = "Featherless"
	ProviderInflection     ProviderName = "Inflection"
	ProviderXAI            ProviderName = "xAI"
	ProviderCloudflare     ProviderName = "Cloudflare"
	Provider01AI           ProviderName = "01.AI"
	ProviderHuggingFace    ProviderName = "HuggingFace"
	ProviderMancer         ProviderName = "Mancer"
	ProviderMancer2        ProviderName = "Mancer 2"
	ProviderHyperbolic     ProviderName = "Hyperbolic"
	ProviderHyperbolic2    ProviderName = "Hyperbolic 2"
	ProviderLynn2          ProviderName = "Lynn 2"
	ProviderLynn           ProviderName = "Lynn"
	ProviderReflection     ProviderName = "Reflection"
)
