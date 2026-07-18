package rbac

import (
	"strings"
)

// MaxMenuIconBytes is the max raw image size for menu icons (32KB).
const MaxMenuIconBytes = 32 * 1024

// SplitPermission splits "{menu_or_feature_full_code_prefix}:{action}" — for feature
// full_codes the whole string is already the permission (menuCode:featureCode).
// Format: exactly one ':' separating resource code and action/feature code.
func SplitPermission(code string) (resource, action string, ok bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", "", false
	}
	i := strings.LastIndex(code, ":")
	if i <= 0 || i == len(code)-1 {
		return "", "", false
	}
	if strings.Count(code, ":") != 1 {
		return "", "", false
	}
	return code[:i], code[i+1:], true
}

// FeatureFullCode builds a feature full_code from menu code + feature code.
func FeatureFullCode(menuCode, featureCode string) string {
	return menuCode + ":" + featureCode
}

// ValidCode reports whether a resource code is non-empty and contains no '.'.
func ValidCode(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	return !strings.Contains(code, ".")
}

// HasPermission reports membership in a permission set.
func HasPermission(set map[string]struct{}, code string) bool {
	_, ok := set[code]
	return ok
}

// ToSet converts a list to a set.
func ToSet(codes []string) map[string]struct{} {
	set := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		if c == "" {
			continue
		}
		set[c] = struct{}{}
	}
	return set
}
