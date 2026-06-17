package github

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/cybozu/neco-containers/sbomreports-github-syncer/internal/render"

	gogithub "github.com/google/go-github/v88/github"
)

type Client struct {
	client *gogithub.Client
}

type CommitRequest struct {
	Owner         string
	Repo          string
	Branch        string
	Message       string
	Files         []render.File
	DeleteMissing bool
	PathPrefix    string
}

type CommitResult struct {
	Skipped   bool
	CommitSHA string
	CommitURL string
}

func NewClient(baseURL, token string) (*Client, error) {
	opts := []gogithub.ClientOptionsFunc{
		gogithub.WithAuthToken(token),
		gogithub.WithTimeout(60 * time.Second),
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" && strings.TrimRight(baseURL, "/") != "https://api.github.com" {
		normalizedBaseURL := strings.TrimRight(baseURL, "/") + "/"
		opts = append(opts, gogithub.WithURLs(&normalizedBaseURL, nil))
	}

	client, err := gogithub.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create GitHub client: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) CommitFiles(ctx context.Context, req CommitRequest) (*CommitResult, error) {
	if len(req.Files) == 0 && !req.DeleteMissing {
		return &CommitResult{Skipped: true}, nil
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	baseRef, _, err := c.client.Git.GetRef(ctx, req.Owner, req.Repo, "heads/"+branch)
	if err != nil {
		return nil, fmt.Errorf("get GitHub ref heads/%s: %w", branch, err)
	}
	if baseRef == nil || baseRef.Object == nil || baseRef.Object.SHA == nil || *baseRef.Object.SHA == "" {
		return nil, fmt.Errorf("get GitHub ref heads/%s: empty object SHA", branch)
	}

	baseCommitSHA := *baseRef.Object.SHA

	baseCommit, _, err := c.client.Git.GetCommit(ctx, req.Owner, req.Repo, baseCommitSHA)
	if err != nil {
		return nil, fmt.Errorf("get GitHub commit %s: %w", baseCommitSHA, err)
	}
	if baseCommit == nil || baseCommit.Tree == nil || baseCommit.Tree.SHA == nil || *baseCommit.Tree.SHA == "" {
		return nil, fmt.Errorf("get GitHub commit %s: empty tree SHA", baseCommitSHA)
	}

	baseTreeSHA := *baseCommit.Tree.SHA

	entries := make([]*gogithub.TreeEntry, 0, len(req.Files))
	generatedPaths := map[string]struct{}{}

	for _, file := range req.Files {
		p, err := cleanRepoPath(file.Path)
		if err != nil {
			return nil, err
		}
		if p == "" {
			return nil, fmt.Errorf("empty repository path generated")
		}

		generatedPaths[p] = struct{}{}
		entries = append(entries, &gogithub.TreeEntry{
			Path:    gogithub.Ptr(p),
			Mode:    gogithub.Ptr("100644"),
			Type:    gogithub.Ptr("blob"),
			Content: gogithub.Ptr(string(file.Content)),
		})
	}

	if req.DeleteMissing {
		prefix, err := cleanRepoPath(req.PathPrefix)
		if err != nil {
			return nil, err
		}
		if prefix == "" {
			return nil, fmt.Errorf("delete missing requires non-empty path prefix")
		}

		existingTree, _, err := c.client.Git.GetTree(ctx, req.Owner, req.Repo, baseTreeSHA, true)
		if err != nil {
			return nil, fmt.Errorf("get recursive tree %s: %w", baseTreeSHA, err)
		}
		if existingTree != nil && existingTree.Truncated != nil && *existingTree.Truncated {
			return nil, fmt.Errorf("GitHub returned a truncated tree; refusing to delete missing files under %q", prefix)
		}

		prefixWithSlash := prefix + "/"
		if existingTree != nil {
			for _, item := range existingTree.Entries {
				if item == nil || item.Path == nil || item.Type == nil {
					continue
				}

				itemPath := *item.Path
				if *item.Type != "blob" || !strings.HasSuffix(itemPath, ".json") {
					continue
				}
				if itemPath != prefix && !strings.HasPrefix(itemPath, prefixWithSlash) {
					continue
				}
				if _, ok := generatedPaths[itemPath]; ok {
					continue
				}

				// go-github serializes TreeEntry with nil SHA and nil Content
				// as {"sha": null}, which GitHub interprets as deletion.
				entries = append(entries, &gogithub.TreeEntry{
					Path: gogithub.Ptr(itemPath),
				})
			}
		}
	}

	newTree, _, err := c.client.Git.CreateTree(ctx, req.Owner, req.Repo, baseTreeSHA, entries)
	if err != nil {
		return nil, fmt.Errorf("create GitHub tree: %w", err)
	}
	if newTree == nil || newTree.SHA == nil || *newTree.SHA == "" {
		return nil, fmt.Errorf("create GitHub tree: empty tree SHA")
	}

	newTreeSHA := *newTree.SHA
	if newTreeSHA == baseTreeSHA {
		return &CommitResult{Skipped: true}, nil
	}

	newCommit, _, err := c.client.Git.CreateCommit(ctx, req.Owner, req.Repo, gogithub.Commit{
		Message: gogithub.Ptr(req.Message),
		Tree: &gogithub.Tree{
			SHA: gogithub.Ptr(newTreeSHA),
		},
		Parents: []*gogithub.Commit{
			{SHA: gogithub.Ptr(baseCommitSHA)},
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("create GitHub commit: %w", err)
	}
	if newCommit == nil || newCommit.SHA == nil || *newCommit.SHA == "" {
		return nil, fmt.Errorf("create GitHub commit: empty commit SHA")
	}

	newCommitSHA := *newCommit.SHA

	_, _, err = c.client.Git.UpdateRef(ctx, req.Owner, req.Repo, "heads/"+branch, gogithub.UpdateRef{
		SHA:   newCommitSHA,
		Force: gogithub.Ptr(false),
	})
	if err != nil {
		return nil, fmt.Errorf("update GitHub ref heads/%s: %w", branch, err)
	}

	result := &CommitResult{
		CommitSHA: newCommitSHA,
	}
	if newCommit.HTMLURL != nil {
		result.CommitURL = *newCommit.HTMLURL
	}

	return result, nil
}

func cleanRepoPath(p string) (string, error) {
	raw := strings.TrimSpace(p)
	raw = strings.Trim(raw, "/")
	if raw == "" || raw == "." {
		return "", nil
	}

	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid repository path %q: dot path segments are not allowed", p)
		}
	}

	cleaned := path.Clean(raw)
	if cleaned == "." {
		return "", nil
	}
	return cleaned, nil
}
