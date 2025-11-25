package domain

import (
	"fmt"
	"regexp"
)

// Locale represents a language/region code for email templates.
type Locale string

// Common locales
const (
	LocaleEnglish    Locale = "en"
	LocaleEnglishUS  Locale = "en-US"
	LocaleEnglishGB  Locale = "en-GB"
	LocaleSpanish    Locale = "es"
	LocaleFrench     Locale = "fr"
	LocaleGerman     Locale = "de"
	LocalePortuguese Locale = "pt"
	LocaleItalian    Locale = "it"
	LocaleDutch      Locale = "nl"
	LocaleJapanese   Locale = "ja"
	LocaleChinese    Locale = "zh"
	LocaleKorean     Locale = "ko"
)

// localePattern matches ISO 639-1 (2 letter) or ISO 639-1 + ISO 3166-1 (e.g., en-US)
var localePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

// String returns the string representation of the locale.
func (l Locale) String() string {
	return string(l)
}

// Validate validates the locale format.
func (l Locale) Validate() error {
	if l == "" {
		return fmt.Errorf("locale cannot be empty")
	}
	if !localePattern.MatchString(string(l)) {
		return fmt.Errorf("invalid locale format: %s (expected format: 'en' or 'en-US')", l)
	}
	return nil
}

// IsValid returns true if the locale is valid.
func (l Locale) IsValid() bool {
	return l.Validate() == nil
}

// Language returns the language part of the locale (e.g., "en" from "en-US").
func (l Locale) Language() string {
	if len(l) >= 2 {
		return string(l)[:2]
	}
	return string(l)
}

// ParseLocale parses a string into a Locale.
func ParseLocale(s string) (Locale, error) {
	locale := Locale(s)
	if err := locale.Validate(); err != nil {
		return "", err
	}
	return locale, nil
}

// DefaultLocale returns the default locale.
func DefaultLocale() Locale {
	return LocaleEnglish
}
