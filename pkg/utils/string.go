package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func Slugify(s string) string {
	orig := strings.TrimSpace(s)
	s = strings.ToLower(orig)
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "-")
	reg2 := regexp.MustCompile(`-+`)
	s = reg2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" && orig != "" {
		// 非拉丁字符（如中文）无法生成 slug 时，用内容哈希兜底保证唯一性
		h := sha1.Sum([]byte(orig))
		s = "u-" + hex.EncodeToString(h[:])[:12]
	}
	return s
}

func IsValidEmail(email string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return reg.MatchString(email)
}

func IsValidPassword(pwd string) bool {
	if len(pwd) < 6 || len(pwd) > 32 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range pwd {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func ContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func UniqueStrings(slice []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(slice))
	for _, s := range slice {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
