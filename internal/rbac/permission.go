package rbac

import (
	"strings"
)

// OpsPathPrefix is the hard-gated ops domain (ops and ops.*).
const OpsPathPrefix = "ops"

// MaxMenuIconBytes is the max raw image size for menu icons (32KB).
const MaxMenuIconBytes = 32 * 1024

// SplitPermission splits "{path}:action" into path and action.
func SplitPermission(code string) (path, action string, ok bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", "", false
	}
	i := strings.LastIndex(code, ":")
	if i <= 0 || i == len(code)-1 {
		return "", "", false
	}
	// Exactly one ':' in the whole string per DESIGN.
	if strings.Count(code, ":") != 1 {
		return "", "", false
	}
	return code[:i], code[i+1:], true
}

// PermissionCode builds "{path}:action".
func PermissionCode(path, action string) string {
	return path + ":" + action
}

// IsOpsPath reports whether path is the ops domain (ops or ops.*).
func IsOpsPath(path string) bool {
	path = strings.TrimSpace(path)
	return path == OpsPathPrefix || strings.HasPrefix(path, OpsPathPrefix+".")
}

// IsOpsPermission reports whether a permission code targets the ops domain.
func IsOpsPermission(code string) bool {
	path, _, ok := SplitPermission(code)
	if !ok {
		return IsOpsPath(code)
	}
	return IsOpsPath(path)
}

// FilterOpsPermissions removes ops-domain codes (for non-super-admin effective sets).
func FilterOpsPermissions(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		if IsOpsPermission(c) {
			continue
		}
		out = append(out, c)
	}
	return out
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
