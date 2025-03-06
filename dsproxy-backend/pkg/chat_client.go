package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const APIBaseURL = "https://api.ppinfra.com/v3/openai"

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
}

type ResponseFormat struct {
	Type   string           `json:"type"`
	Schema *json.RawMessage `json:"schema,omitempty"`
}

type ChatCompletionRequest struct {
	Model             string           `json:"model"`
	Messages          []RequestMessage `json:"messages"`
	MaxTokens         uint32           `json:"max_tokens"`
	Temperature       *float32         `json:"temperature,omitempty"`
	TopP              *float32         `json:"top_p,omitempty"`
	TopK              *uint32          `json:"top_k,omitempty"`
	MinP              *float32         `json:"min_p,omitempty"`
	N                 *uint32          `json:"n,omitempty"`
	Stream            *bool            `json:"stream,omitempty"`
	StreamOptions     *StreamOptions   `json:"stream_options,omitempty"`
	Stop              []string         `json:"stop,omitempty"`
	PresencePenalty   *float32         `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float32         `json:"frequency_penalty,omitempty"`
	RepetitionPenalty *float32         `json:"repetition_penalty,omitempty"`
	LogitBias         map[string]int32 `json:"logit_bias,omitempty"`
	Logprobs          *bool            `json:"logprobs,omitempty"`
	TopLogprobs       *uint32          `json:"top_logprobs,omitempty"`
	ResponseFormat    *ResponseFormat  `json:"response_format,omitempty"`
	Seed              *uint32          `json:"seed,omitempty"`
	User              *string          `json:"user,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Usage struct {
	PromptTokens     uint32 `json:"prompt_tokens"`
	CompletionTokens uint32 `json:"completion_tokens"`
	TotalTokens      uint32 `json:"total_tokens"`
}

// Delta represents incremental content in streaming responses
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamChatChoice represents a choice in streaming responses
type StreamChatChoice struct {
	Index                uint32                 `json:"index"`
	Delta                Delta                  `json:"delta"`
	FinishReason         *string                `json:"finish_reason"` // Nullable
	ContentFilterResults map[string]interface{} `json:"content_filter_results,omitempty"`
}

// StreamChatCompletionResponse represents a single chunk in streaming responses
type StreamChatCompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           uint64             `json:"created"`
	Model             string             `json:"model"`
	Choices           []StreamChatChoice `json:"choices"`
	SystemFingerprint string             `json:"system_fingerprint"`
	Usage             *Usage             `json:"usage,omitempty"` // Optional, only in last chunk
}

type ChatClient struct {
	client *http.Client
	apiKey string
}

func NewChatClient(apiKey string) *ChatClient {
	return &ChatClient{
		client: &http.Client{},
		apiKey: apiKey,
	}
}

func (c *ChatClient) post(endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (c *ChatClient) CreateChatCompletionStream(request ChatCompletionRequest, handler func(*StreamChatCompletionResponse) error) error {
	resp, err := c.post("chat/completions", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		if line == "data: [DONE]" {
			break
		}

		jsonData := line[6:] // Remove "data: " prefix
		var streamResp StreamChatCompletionResponse
		if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
			log.Printf("Failed to unmarshal stream chunk: %v", err)
			continue
		}

		if err := handler(&streamResp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}
