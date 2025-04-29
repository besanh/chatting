package service

import googletranslate "github.com/besanh/chatting/external/google"

type (
	ISubscriber interface {
	}

	Subscriber struct {
		googleTranslate googletranslate.IGoogleTranslate
	}
)
