package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Operation is one extracted bank operation. Mirrors services.BankOperation
// but lives here to keep the ai package dependency-free of services.
type Operation struct {
	Date          string  `json:"date"`
	Amount        float64 `json:"amount"`
	PayerINN      string  `json:"payer_inn"`
	PayerName     string  `json:"payer_name"`
	Purpose       string  `json:"purpose"`
	InvoiceNumber string  `json:"invoice_number"`
}

// Extractor turns raw statement text into structured operations.
type Extractor interface {
	Extract(ctx context.Context, rawText string) ([]Operation, error)
	Name() string
}

// Config controls which extractor is built.
type Config struct {
	Provider string // "", "openai" (or any OpenAI-compatible), "ollama"
	BaseURL  string
	APIKey   string
	Model    string
	Log      *slog.Logger
}

// NewExtractor returns an LLM-backed extractor when configured, otherwise a
// regex fallback that needs no API keys/tokens. This lets the feature ship and
// run with zero external dependencies, while keeping the AI path one env away.
func NewExtractor(cfg Config) Extractor {
	switch strings.ToLower(cfg.Provider) {
	case "openai", "ollama":
		base := cfg.BaseURL
		if base == "" {
			if cfg.Provider == "ollama" {
				base = "http://localhost:11434/v1"
			} else {
				base = "https://api.openai.com/v1"
			}
		}
		cfg.Log.Info("AI extractor enabled", "provider", cfg.Provider, "base_url", base, "model", cfg.Model)
		return &openAIExtractor{
			baseURL: strings.TrimRight(base, "/"),
			apiKey:  cfg.APIKey,
			model:   cfg.Model,
			log:     cfg.Log,
			client:  &http.Client{Timeout: 30 * time.Second},
		}
	default:
		cfg.Log.Info("AI extractor disabled, using regex fallback (no API key required)")
		return &regexExtractor{}
	}
}

// ---- OpenAI-compatible extractor ----

type openAIExtractor struct {
	baseURL string
	apiKey  string
	model   string
	log     *slog.Logger
	client  *http.Client
}

func (e *openAIExtractor) Name() string { return "llm:" + e.model }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	ResponseFormat *respFormat   `json:"response_format,omitempty"`
}

type respFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (e *openAIExtractor) Extract(ctx context.Context, rawText string) ([]Operation, error) {
	reqBody := chatRequest{
		Model:       e.model,
		Temperature: 0, // deterministic extraction
		Messages: []chatMessage{
			{Role: "system", Content: SystemPrompt + "\n\n" + FewShotExample},
			{Role: "user", Content: fmt.Sprintf(UserPromptTemplate, rawText)},
		},
		// Ask for JSON mode where supported (OpenAI, many proxies). Harmless if ignored.
		ResponseFormat: &respFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm returned %d: %s", resp.StatusCode, truncate(string(body), 300))
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}
	if cr.Error != nil {
		return nil, fmt.Errorf("llm error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("llm returned no choices")
	}

	content := stripCodeFence(cr.Choices[0].Message.Content)

	var parsed struct {
		Operations []Operation `json:"operations"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("llm output is not valid JSON: %w (content: %s)", err, truncate(content, 200))
	}

	e.log.Info("llm extraction succeeded", "operations", len(parsed.Operations))
	return parsed.Operations, nil
}

// stripCodeFence removes ```json ... ``` wrappers some models add despite instructions.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ---- Regex fallback extractor ----
//
// Deterministic, dependency-free baseline. Handles the common Russian bank
// statement shape used in the seed/sample. Not as flexible as an LLM, but
// works offline with zero configuration — the default path.

type regexExtractor struct{}

func (r *regexExtractor) Name() string { return "regex" }

var (
	reDate    = regexp.MustCompile(`(\d{2})[.\-/](\d{2})[.\-/](\d{4})`)
	reAmount  = regexp.MustCompile(`([\d\s]+[.,]\d{2})\s*(?:RUB|руб|₽)?`)
	reINN     = regexp.MustCompile(`ИНН[:\s]*(\d{10,12})`)
	reInvoice = regexp.MustCompile(`(?:счёт|счет|сч\.?)\s*№?\s*(\d+)`)
	rePayer   = regexp.MustCompile(`(?:Плательщик|Отправитель)[:\s]*([^,\n]+)`)
	rePurpose = regexp.MustCompile(`(?:Назначение|Назначение платежа)[:\s]*([^\n]+)`)
)

func (r *regexExtractor) Extract(_ context.Context, rawText string) ([]Operation, error) {
	// Split into blocks separated by blank lines; treat each block as one op.
	blocks := regexp.MustCompile(`\n\s*\n`).Split(rawText, -1)
	var ops []Operation

	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		var op Operation

		if m := reDate.FindStringSubmatch(block); m != nil {
			op.Date = fmt.Sprintf("%s-%s-%s", m[3], m[2], m[1]) // DD.MM.YYYY -> YYYY-MM-DD
		}
		if m := reAmount.FindStringSubmatch(block); m != nil {
			clean := strings.ReplaceAll(m[1], " ", "")
			clean = strings.ReplaceAll(clean, ",", ".")
			op.Amount, _ = strconv.ParseFloat(clean, 64)
		}
		if m := reINN.FindStringSubmatch(block); m != nil {
			op.PayerINN = m[1]
		}
		if m := rePayer.FindStringSubmatch(block); m != nil {
			op.PayerName = strings.TrimSpace(m[1])
		}
		if m := rePurpose.FindStringSubmatch(block); m != nil {
			op.Purpose = strings.TrimSpace(m[1])
		}
		if m := reInvoice.FindStringSubmatch(block); m != nil {
			op.InvoiceNumber = m[1]
		}

		// Only keep blocks that look like a real payment.
		if op.Date != "" && op.Amount > 0 {
			ops = append(ops, op)
		}
	}

	return ops, nil
}
