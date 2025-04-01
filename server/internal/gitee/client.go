package gitee

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client 码云API客户端
type Client struct {
	token      string
	httpClient *http.Client
}

// Repository 仓库信息
type Repository struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	HTMLURL     string    `json:"html_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewClient 创建码云API客户端
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRepoFromURL 从URL获取码云仓库信息
func (c *Client) GetRepoFromURL(repoURL string) (*Repository, error) {
	owner, repoName, err := parseRepoURL(repoURL)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s", owner, repoName)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Add("Authorization", "token "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("仓库不存在或无访问权限")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("码云API返回错误: %d", resp.StatusCode)
	}

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

// parseRepoURL 从URL中提取所有者和仓库名
func parseRepoURL(repoURL string) (string, string, error) {
	// 处理URL格式：https://gitee.com/owner/repo.git 或 https://gitee.com/owner/repo
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}

	if u.Host != "gitee.com" {
		return "", "", errors.New("仅支持码云仓库URL")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", errors.New("无效的仓库URL格式")
	}

	owner := parts[0]
	repo := parts[1]
	repo = strings.TrimSuffix(repo, ".git")

	return owner, repo, nil
}