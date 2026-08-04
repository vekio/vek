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

type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func DefaultTemplateClient() HTTPClient {
	return HTTPClient{
		BaseURL: defaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c HTTPClient) GetTemplate(ctx context.Context, name string) (Template, error) {
	name = NormalizeTemplateName(name)
	if err := validateTemplateName(name); err != nil {
		return Template{}, err
	}

	requestURL, err := c.templateURL(name)
	if err != nil {
		return Template{}, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = DefaultTemplateClient().HTTPClient
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return Template{}, fmt.Errorf("build request for %q: %w", name, err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return Template{}, fmt.Errorf("download gitignore template %q: %w", name, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return Template{}, fmt.Errorf("gitignore template %q not found", name)
	}

	if response.StatusCode != http.StatusOK {
		return Template{}, fmt.Errorf("download gitignore template %q: unexpected status %s", name, response.Status)
	}

	content, err := io.ReadAll(io.LimitReader(response.Body, maxTemplateSize+1))
	if err != nil {
		return Template{}, fmt.Errorf("read gitignore template %q: %w", name, err)
	}

	if len(content) > maxTemplateSize {
		return Template{}, fmt.Errorf("gitignore template %q is too large", name)
	}

	return Template{
		Name:    name,
		Content: string(content),
	}, nil
}

func (c HTTPClient) templateURL(name string) (string, error) {
	baseURL := strings.TrimRight(c.BaseURL, "/")
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
