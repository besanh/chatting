package googletranslate

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/besanh/chatting/common/util"
	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/model"
	log "github.com/besanh/logger/logging/slog"
	"resty.dev/v3"
)

type IGoogleTranslate interface {
	GetGoogleTranslateApi(request model.GoogleTranslateApiRequest) (*model.TranslationResponse, error)
	PostGoogleTranslateApi(url, key string, body []any) (*model.TranslationResponse, error)
}

type GoogleTranslate struct {
	config config.Config
}

func NewGoogleTranslate(cfg config.Config) IGoogleTranslate {
	return &GoogleTranslate{config: cfg}
}

func (g *GoogleTranslate) GetGoogleTranslateApi(request model.GoogleTranslateApiRequest) (*model.TranslationResponse, error) {
	result := &model.TranslationResponse{}
	query := map[string]string{}
	if v := request.Client; v != "" {
		query["client"] = v
	}
	if v := request.Sl; v != "" {
		query["sl"] = v
	}
	if v := request.Tl; v != "" {
		query["tl"] = v
	}
	if v := request.Dt; v != "" {
		query["dt"] = v
	}
	if v := request.Q; v != "" {
		query["q"] = v
	}

	client := resty.New()
	defer client.Close()

	resp, err := client.R().SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   util.RandomChoice(util.UserAgents),
	}).SetQueryParams(query).Get(request.Url)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode())
	}

	var content string
	switch {
	case strings.Contains(request.Url, "/translate_a/single"):
		content, err = parseSingleSegments(resp.Bytes())
	case strings.Contains(request.Url, "/translate_a/t"):
		content, err = parseSimpleText(resp.Bytes())
	default:
		content, err = extractAndFlatten(resp.Bytes())
	}
	if err != nil {
		return nil, err
	}
	result.Content = html.UnescapeString(content)
	return result, nil
}

func (g *GoogleTranslate) PostGoogleTranslateApi(url, key string, body []any) (*model.TranslationResponse, error) {
	result := &model.TranslationResponse{}
	client := resty.New()
	defer client.Close()

	resp, err := client.R().SetHeaders(map[string]string{
		"Content-Type":   "application/json+protobuf",
		"Accept":         "application/json+protobuf",
		"X-Goog-API-Key": key,
		"User-Agent":     util.RandomChoice(util.UserAgents),
	}).SetBody(body).Post(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode())
	}

	var content string
	switch {
	case strings.Contains(url, "translateHtml"):
		content, err = parseSimpleText(resp.Bytes())
	default:
		content, err = extractAndFlatten(resp.Bytes())
	}
	if err != nil {
		return nil, err
	}
	result.Content = html.UnescapeString(content)
	return result, nil
}

func parseSingleSegments(data []byte) (result string, err error) {
	var arr []any
	if err = json.Unmarshal(data, &arr); err != nil {
		err = fmt.Errorf("invalid JSON for single segments: %w", err)
		return
	}
	if len(arr) == 0 {
		err = fmt.Errorf("empty response array")
		return
	}
	segments, ok := arr[0].([]any)
	if !ok {
		err = fmt.Errorf("unexpected format for single segments: %T", arr[0])
		return
	}
	var sb strings.Builder
	for _, seg := range segments {
		if s, ok := seg.([]any); ok && len(s) > 0 {
			if txt, ok := s[0].(string); ok {
				sb.WriteString(txt)
			}
		}
	}
	result = sb.String()
	if result == "" {
		err = fmt.Errorf("no translation found in single segments")
		return
	}
	return
}

func parseSimpleText(data []byte) (result string, err error) {
	var arr []any
	if err = json.Unmarshal(data, &arr); err != nil {
		err = fmt.Errorf("invalid JSON for simple response: %w", err)
		return
	}
	if len(arr) == 0 {
		err = fmt.Errorf("empty response array")
		return
	}
	first, ok := arr[0].([]any)
	if !ok || len(first) == 0 {
		err = fmt.Errorf("unexpected format for simple response: %T", arr[0])
		return
	}
	if result, ok = first[0].(string); ok {
		return
	}

	err = fmt.Errorf("no translation found in simple response")

	return
}

func extractAndFlatten(data []byte) (result string, err error) {
	var raw any
	if err = json.Unmarshal(data, &raw); err != nil {
		err = fmt.Errorf("invalid JSON: %w", err)
		return
	}
	collected := flatten(raw)
	if len(collected) == 0 {
		err = fmt.Errorf("no translation strings found")
		return
	}
	result = strings.Join(collected, "")

	return
}

func flatten(v any) []string {
	switch x := v.(type) {
	case string:
		return []string{x}
	case []any:
		var out []string
		for _, e := range x {
			out = append(out, flatten(e)...)
		}
		return out
	default:
		return nil
	}
}
