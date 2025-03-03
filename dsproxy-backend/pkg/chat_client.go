package pkg

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
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
    Type   string         `json:"type"`
    Schema *json.RawMessage `json:"schema,omitempty"`
}

type ChatCompletionRequest struct {
    Model            string            `json:"model"`
    Messages         []RequestMessage  `json:"messages"`
    MaxTokens        uint32            `json:"max_tokens"`
    Temperature      *float32          `json:"temperature,omitempty"`
    TopP             *float32          `json:"top_p,omitempty"`
    TopK             *uint32           `json:"top_k,omitempty"`
    MinP             *float32          `json:"min_p,omitempty"`
    N                *uint32           `json:"n,omitempty"`
    Stream           *bool             `json:"stream,omitempty"`
    Stop             []string          `json:"stop,omitempty"`
    PresencePenalty  *float32          `json:"presence_penalty,omitempty"`
    FrequencyPenalty *float32          `json:"frequency_penalty,omitempty"`
    RepetitionPenalty *float32         `json:"repetition_penalty,omitempty"`
    LogitBias        map[string]int32  `json:"logit_bias,omitempty"`
    Logprobs         *bool             `json:"logprobs,omitempty"`
    TopLogprobs      *uint32           `json:"top_logprobs,omitempty"`
    ResponseFormat   *ResponseFormat   `json:"response_format,omitempty"`
    Seed             *uint32           `json:"seed,omitempty"`
    User             *string           `json:"user,omitempty"`
}

type ChatChoice struct {
    Index        uint32         `json:"index"`
    Message      ResponseMessage `json:"message"`
    FinishReason string         `json:"finish_reason"`
}

type Usage struct {
    PromptTokens     uint32 `json:"prompt_tokens"`
    CompletionTokens uint32 `json:"completion_tokens"`
    TotalTokens      uint32 `json:"total_tokens"`
}

type ChatCompletionResponse struct {
    ID      string       `json:"id"`
    Object  string       `json:"object"`
    Created uint64       `json:"created"`
    Model   string       `json:"model"`
    Choices []ChatChoice `json:"choices"`
    Usage   *Usage       `json:"usage,omitempty"`
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

// CreateChatCompletion handles non-streaming responses
func (c *ChatClient) CreateChatCompletion(request ChatCompletionRequest) (*ChatCompletionResponse, error) {
    resp, err := c.post("chat/completions", request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }

    var response ChatCompletionResponse
    err = json.Unmarshal(body, &response)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %v", err)
    }

    return &response, nil
}

// CreateChatCompletionStream handles streaming responses
func (c *ChatClient) CreateChatCompletionStream(request ChatCompletionRequest, handler func(ChatCompletionResponse) error) error {
    // Ensure stream is set to true
    streamTrue := true
    request.Stream = &streamTrue

    resp, err := c.post("chat/completions", request)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        // Skip empty lines or non-data lines
        if line == "" || !strings.HasPrefix(line, "data: ") {
            continue
        }

        // Check for stream end
        if line == "data: [DONE]" {
            break
        }

        // Parse the JSON data
        jsonData := line[6:] // Remove "data: " prefix
        var response ChatCompletionResponse
        err = json.Unmarshal([]byte(jsonData), &response)
        if err != nil {
            return fmt.Errorf("failed to unmarshal stream chunk: %v", err)
        }

        // Call the handler with the parsed response
        if err := handler(response); err != nil {
            return err
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading stream: %v", err)
    }

    return nil
}