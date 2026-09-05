package models

import "strings"

// CurrencyOption is one allowed group currency for API docs and the UI.
type CurrencyOption struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Country string `json:"country"` // ISO 3166-1 alpha-2 (or "eu") for flag images
}

// Currencies is the fixed set accepted by POST /groups.
var Currencies = []CurrencyOption{
	{Code: "USD", Name: "US Dollar", Country: "us"},
	{Code: "EUR", Name: "Euro", Country: "eu"},
	{Code: "GBP", Name: "British Pound", Country: "gb"},
	{Code: "CAD", Name: "Canadian Dollar", Country: "ca"},
	{Code: "AUD", Name: "Australian Dollar", Country: "au"},
	{Code: "JPY", Name: "Japanese Yen", Country: "jp"},
	{Code: "CNY", Name: "Chinese Yuan", Country: "cn"},
	{Code: "CHF", Name: "Swiss Franc", Country: "ch"},
	{Code: "INR", Name: "Indian Rupee", Country: "in"},
	{Code: "MXN", Name: "Mexican Peso", Country: "mx"},
	{Code: "BRL", Name: "Brazilian Real", Country: "br"},
	{Code: "KRW", Name: "South Korean Won", Country: "kr"},
}

// DefaultCurrency is used when the client omits currency.
const DefaultCurrency = "USD"

// NormalizeCurrency uppercases and trims; empty becomes DefaultCurrency.
func NormalizeCurrency(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return DefaultCurrency
	}
	return code
}

// IsAllowedCurrency reports whether code is in Currencies.
func IsAllowedCurrency(code string) bool {
	code = NormalizeCurrency(code)
	for _, c := range Currencies {
		if c.Code == code {
			return true
		}
	}
	return false
}

// CurrencyCodes returns just the ISO codes (for Swagger enum / docs).
func CurrencyCodes() []string {
	out := make([]string, len(Currencies))
	for i, c := range Currencies {
		out[i] = c.Code
	}
	return out
}
