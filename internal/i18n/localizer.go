package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	LanguageEnglish  = "en"
	LanguageGeorgian = "ka"
	DefaultLanguage  = LanguageEnglish
)

//go:embed locales/*.json
var localeFiles embed.FS

type Localizer struct {
	messages map[string]map[string]string
}

// NewLocalizer loads embedded message catalogs for all supported languages.
func NewLocalizer() (*Localizer, error) {
	languages := []string{LanguageEnglish, LanguageGeorgian}
	messages := make(map[string]map[string]string, len(languages))

	for _, language := range languages {
		payload, err := localeFiles.ReadFile(fmt.Sprintf("locales/%s.json", language))
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", language, err)
		}

		var catalog map[string]string
		if err := json.Unmarshal(payload, &catalog); err != nil {
			return nil, fmt.Errorf("unmarshal locale %s: %w", language, err)
		}

		messages[language] = catalog
	}

	return &Localizer{messages: messages}, nil
}

// MustNewLocalizer returns a configured localizer or panics if catalogs cannot be loaded.
func MustNewLocalizer() *Localizer {
	localizer, err := NewLocalizer()
	if err != nil {
		panic(err)
	}

	return localizer
}

// Resolve picks the first supported language from an Accept-Language header value.
func (l *Localizer) Resolve(header string) string {
	for part := range strings.SplitSeq(header, ",") {
		language := normalizeLanguage(part)
		if language == "" {
			continue
		}

		if _, ok := l.messages[language]; ok {
			return language
		}
	}

	return DefaultLanguage
}

// Message returns a localized message for the given key and header value.
func (l *Localizer) Message(header string, key string) string {
	if l == nil {
		return key
	}

	language := l.Resolve(header)
	if message, ok := l.messages[language][key]; ok {
		return message
	}

	if message, ok := l.messages[DefaultLanguage][key]; ok {
		return message
	}

	return key
}

// normalizeLanguage extracts a base language code from an Accept-Language item.
func normalizeLanguage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if index := strings.IndexByte(value, ';'); index >= 0 {
		value = value[:index]
	}

	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	if index := strings.IndexAny(value, "-_"); index >= 0 {
		value = value[:index]
	}

	return value
}
