// Package modelname formats raw model identifiers into human-friendly display
// names, matching the "Title Case + spaces" convention used by the built-in
// model catalogs (e.g. "claude-sonnet-5" -> "Claude Sonnet 5").
//
// It is used as a fallback when a model id is not found in a platform's curated
// DefaultModels list — typically models freshly pulled from an upstream via the
// "sync upstream models" feature — so the admin test-model dropdown stays
// visually consistent instead of showing raw lowercase-hyphenated ids.
package modelname

import (
	"strings"
	"unicode"
)

// acronyms maps lowercased tokens that should render in a fixed, non-title form.
var acronyms = map[string]string{
	"gpt":      "GPT",
	"ai":       "AI",
	"api":      "API",
	"tts":      "TTS",
	"hd":       "HD",
	"glm":      "GLM",
	"deepseek": "DeepSeek",
}

// FormatDisplayName converts a raw model id such as "claude-sonnet-5" into a
// human-friendly display name such as "Claude Sonnet 5".
//
// Rules (best-effort; never panics, and returns a non-empty result whenever the
// input contains non-space characters):
//   - tokens are split on "-"
//   - a trailing date suffix is dropped, matching the built-in catalog
//     convention — both the compact "...-20251101" form ("claude-opus-4-5-20251101"
//     -> "Claude Opus 4.5") and the dashed "...-2024-04-09" form
//     ("gpt-4-turbo-2024-04-09" -> "GPT 4 Turbo")
//   - runs of short numeric tokens (1-2 digits) are joined with "." so version
//     pairs render as "4.6" / "3.5" rather than "4 6" / "3 5"
//   - alphabetic tokens are Title-cased, with a few well-known acronyms uppercased
//
// If the input is empty (or whitespace-only) it is returned unchanged.
func FormatDisplayName(id string) string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return id
	}

	tokens := dropTrailingDate(strings.Split(trimmed, "-"))

	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "" {
			continue
		}

		// Coalesce a run of short numeric tokens into a dotted version string.
		if isShortNumeric(tok) {
			parts := []string{tok}
			for i+1 < len(tokens) && isShortNumeric(tokens[i+1]) {
				parts = append(parts, tokens[i+1])
				i++
			}
			out = append(out, strings.Join(parts, "."))
			continue
		}

		out = append(out, titleToken(tok))
	}

	if len(out) == 0 {
		return trimmed
	}
	return strings.Join(out, " ")
}

// titleToken capitalizes an alphabetic token (or maps a known acronym), while
// leaving tokens that start with a digit (e.g. "4o", "5.4") untouched.
func titleToken(tok string) string {
	if repl, ok := acronyms[strings.ToLower(tok)]; ok {
		return repl
	}
	runes := []rune(tok)
	if len(runes) > 0 && unicode.IsLetter(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

// dropTrailingDate removes a trailing date suffix so snapshot ids render like
// their undated catalog counterparts. It handles the compact "20251101" form and
// the dashed "2024-04-09" (YYYY-MM-DD) form, but only when at least one non-date
// token precedes it, and never touches short numeric version pairs like "4-6".
func dropTrailingDate(tokens []string) []string {
	n := len(tokens)
	// Compact YYYYMMDD, e.g. "claude-opus-4-5-20251101".
	if n > 1 && isDate8(tokens[n-1]) {
		return tokens[:n-1]
	}
	// Dashed YYYY-MM-DD, e.g. "gpt-4-turbo-2024-04-09". Requires a name token
	// before the date, and a zero-padded 2-digit month/day (the only form real
	// snapshot ids use) so a genuine single-digit version tail like "…-2024-1-5"
	// is left untouched rather than mistaken for a date.
	if n > 3 {
		mm, dd := tokens[n-2], tokens[n-1]
		if isYear(tokens[n-3]) && len(mm) == 2 && len(dd) == 2 && isMonth(mm) && isDay(dd) {
			return tokens[:n-3]
		}
	}
	return tokens
}

func isDate8(s string) bool {
	return len(s) == 8 && allDigits(s)
}

func isYear(s string) bool {
	return len(s) == 4 && allDigits(s) &&
		(strings.HasPrefix(s, "19") || strings.HasPrefix(s, "20") || strings.HasPrefix(s, "21"))
}

func isMonth(s string) bool {
	n, ok := smallNum(s)
	return ok && n >= 1 && n <= 12
}

func isDay(s string) bool {
	n, ok := smallNum(s)
	return ok && n >= 1 && n <= 31
}

// smallNum parses a 1-2 digit numeric token, reporting whether it qualified.
func smallNum(s string) (int, bool) {
	if len(s) < 1 || len(s) > 2 || !allDigits(s) {
		return 0, false
	}
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n, true
}

func isShortNumeric(s string) bool {
	return len(s) >= 1 && len(s) <= 2 && allDigits(s)
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
