package deployer

import (
	"path"
	"strings"
)

func isWindowsServer(server ServerInfo) bool {
	return strings.EqualFold(server.OSType, "windows")
}

func normalizeRemotePath(server ServerInfo, remotePath string) string {
	trimmed := strings.TrimSpace(remotePath)
	if trimmed == "" {
		return ""
	}
	if isWindowsServer(server) {
		return strings.ReplaceAll(trimmed, "/", "\\")
	}
	return path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
}

func joinRemotePath(server ServerInfo, elements ...string) string {
	if isWindowsServer(server) {
		cleaned := make([]string, 0, len(elements))
		for _, element := range elements {
			part := strings.Trim(strings.ReplaceAll(element, "/", "\\"), "\\")
			if part != "" && part != "." {
				cleaned = append(cleaned, part)
			}
		}
		return strings.Join(cleaned, "\\")
	}

	cleaned := make([]string, 0, len(elements))
	for _, element := range elements {
		part := strings.Trim(strings.ReplaceAll(element, "\\", "/"), "/")
		if part != "" && part != "." {
			cleaned = append(cleaned, part)
		}
	}
	return path.Join(cleaned...)
}

func remoteDir(server ServerInfo, remotePath string) string {
	normalized := normalizeRemotePath(server, remotePath)
	if normalized == "" {
		return ""
	}
	if isWindowsServer(server) {
		index := strings.LastIndex(normalized, `\`)
		if index < 0 {
			return ""
		}
		return normalized[:index]
	}
	return path.Dir(normalized)
}
