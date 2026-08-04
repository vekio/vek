package gitignore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultBaseURL = "https://raw.githubusercontent.com/github/gitignore/main"
const maxTemplateSize = 1 << 20

type httpTemplateClient struct {
	baseURL    string
	httpClient *http.Client
}

func defaultTemplateClient() httpTemplateClient {
	return httpTemplateClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c httpTemplateClient) getTemplate(ctx context.Context, name string) (templateContent, error) {
	name = normalizeTemplateName(name)
	if err := validateTemplateName(name); err != nil {
		return templateContent{}, err
	}
	requestURL, err := c.templateURL(name)
	if err != nil {
		return templateContent{}, err
	}
	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = defaultTemplateClient().httpClient
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return templateContent{}, fmt.Errorf("build request for %q: %w", name, err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return templateContent{}, fmt.Errorf("download gitignore template %q: %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return templateContent{}, fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
	}
	if response.StatusCode != http.StatusOK {
		return templateContent{}, fmt.Errorf("download gitignore template %q: unexpected status %s", name, response.Status)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxTemplateSize+1))
	if err != nil {
		return templateContent{}, fmt.Errorf("read gitignore template %q: %w", name, err)
	}
	if len(content) > maxTemplateSize {
		return templateContent{}, fmt.Errorf("gitignore template %q is too large", name)
	}
	return templateContent{Name: name, Content: string(content)}, nil
}

func (c httpTemplateClient) templateURL(name string) (string, error) {
	baseURL := strings.TrimRight(c.baseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, name+".gitignore")
	return parsedURL.String(), nil
}
