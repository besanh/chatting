package model

import "errors"

type (
	GoogleTranslateApiRequest struct {
		Url    string `json:"url"`
		Client string `json:"client"`
		Sl     string `json:"sl"`
		Tl     string `json:"tl"`
		Dt     string `json:"dt"`
		Q      string `json:"q"`
		Engine string `json:"engine"`
	}
)

func ToNestedJSON(req GoogleTranslateApiRequest) (nested []any) {
	nested = []any{
		[]any{
			[]string{req.Q},
			req.Sl,
			req.Tl,
		},
		req.Engine,
	}

	return
}

type (
	TranslationRequest struct {
		Content   string `json:"content"`
		SessionId string `json:"session_id"`
		LangCode  string `json:"lang_code"`
		MsgID     string `json:"msg_id"`
	}

	TranslationResponse struct {
		Content      string  `json:"content"`
		MsgID        string  `json:"msg_id"`
		DebugContent *string `json:"debug_content,omitempty"`
	}
)

func (m *TranslationRequest) Validate() error {
	if len(m.Content) < 1 {
		return errors.New("content is empty")
	}
	if len(m.SessionId) < 1 {
		return errors.New("session_id is empty")
	}
	if len(m.LangCode) < 1 {
		return errors.New("lang_code is empty")
	}
	if len(m.MsgID) < 1 {
		return errors.New("msg_id is empty")
	}
	return nil
}
