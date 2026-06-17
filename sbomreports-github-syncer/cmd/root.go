package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	githubsync "github.com/cybozu/neco-containers/sbomreports-github-syncer/internal/github"
	"github.com/cybozu/neco-containers/sbomreports-github-syncer/internal/kube"
	"github.com/cybozu/neco-containers/sbomreports-github-syncer/internal/render"
)

type syncOptions struct {
	Kubeconfig string
	Namespace  string
	Selector   string

	GitHubToken  string
	GitHubOwner  string
	GitHubRepo   string
	GitHubBranch string
	GitHubAPIURL string

	ClusterName   string
	PathPrefix    string
	CommitMessage string

	DeleteMissing bool
	FailIfEmpty   bool
	DryRun        bool
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "sbomreports-github-syncer",
		Short:         "Sync Trivy Operator SbomReport resources to a GitHub repository",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	syncOpts := &syncOptions{
		GitHubToken: envOrDefault("GITHUB_TOKEN", ""),
	}
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "List SbomReports from Kubernetes and commit them to GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd.Context(), syncOpts)
		},
	}

	flags := syncCmd.Flags()
	flags.StringVar(&syncOpts.Kubeconfig, "kubeconfig", envOrDefault("KUBECONFIG", ""), "Path to kubeconfig. Empty means in-cluster config first, then ~/.kube/config")
	flags.StringVar(&syncOpts.Namespace, "namespace", envOrDefault("NAMESPACE", ""), "Namespace to list SbomReports from. Empty means all namespaces")
	flags.StringVar(&syncOpts.Selector, "selector", envOrDefault("LABEL_SELECTOR", ""), "Kubernetes label selector for SbomReports")

	flags.StringVar(&syncOpts.GitHubOwner, "github-owner", envOrDefault("GITHUB_OWNER", ""), "GitHub repository owner")
	flags.StringVar(&syncOpts.GitHubRepo, "github-repo", envOrDefault("GITHUB_REPO", ""), "GitHub repository name")
	flags.StringVar(&syncOpts.GitHubBranch, "github-branch", envOrDefault("GITHUB_BRANCH", "main"), "GitHub branch to update")
	flags.StringVar(&syncOpts.GitHubAPIURL, "github-api-url", envOrDefault("GITHUB_API_URL", "https://api.github.com"), "GitHub API URL. For GitHub Enterprise, set the API base URL")

	flags.StringVar(&syncOpts.ClusterName, "cluster-name", envOrDefault("CLUSTER_NAME", ""), "Cluster name included in index metadata")
	flags.StringVar(&syncOpts.PathPrefix, "path-prefix", envOrDefault("GITHUB_PATH_PREFIX", "sbomreports"), "Repository directory prefix for generated files")
	flags.StringVar(&syncOpts.CommitMessage, "commit-message", envOrDefault("COMMIT_MESSAGE", ""), "Commit message. Empty means generated message")

	flags.BoolVar(&syncOpts.DeleteMissing, "delete-missing", envBoolOrDefault("DELETE_MISSING", false), "Delete stale .json files under path-prefix that are not present in the current Kubernetes result")
	flags.BoolVar(&syncOpts.FailIfEmpty, "fail-if-empty", envBoolOrDefault("FAIL_IF_EMPTY", false), "Return an error if no SbomReports are found")
	flags.BoolVar(&syncOpts.DryRun, "dry-run", false, "List and render reports but do not call GitHub")

	root.AddCommand(syncCmd)
	return root
}

func runSync(ctx context.Context, opts *syncOptions) error {
	if opts.DeleteMissing && strings.Trim(opts.PathPrefix, "/") == "" {
		return fmt.Errorf("--delete-missing requires a non-empty --path-prefix to avoid deleting unrelated repository files")
	}
	if !opts.DryRun {
		if opts.GitHubToken == "" {
			return fmt.Errorf("GITHUB_TOKEN is required")
		}
		if opts.GitHubOwner == "" || opts.GitHubRepo == "" {
			return fmt.Errorf("GITHUB_OWNER/GITHUB_REPO or --github-owner/--github-repo are required")
		}
	}

	kubeClient, err := kube.NewDynamicClient(opts.Kubeconfig)
	if err != nil {
		return err
	}

	reports, err := kube.ListSbomReports(ctx, kubeClient, opts.Namespace, opts.Selector)
	if err != nil {
		return err
	}
	if len(reports) == 0 && opts.FailIfEmpty {
		return fmt.Errorf("no SbomReports found")
	}

	nsTeamLabels := map[string]string{}

	for _, report := range reports {
		namespaceName := report.GetNamespace()
		if namespaceName == "" {
			continue
		}

		if _, ok := nsTeamLabels[namespaceName]; ok {
			continue
		}

		labelValue, found, err := kube.GetNamespaceLabel(
			ctx,
			kubeClient,
			namespaceName,
			"team",
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: get team label for namespace %q: %v\n", namespaceName, err)
			continue
		}
		if found {
			nsTeamLabels[namespaceName] = labelValue
		}
	}

	files, err := render.RenderSbomReports(reports, render.Options{
		PathPrefix:          opts.PathPrefix,
		ClusterName:         opts.ClusterName,
		NamespaceTeamLabels: nsTeamLabels,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "found %d SbomReports, rendered %d files\n", len(reports), len(files))
	for _, file := range files {
		fmt.Fprintf(os.Stderr, "  %s (%d bytes)\n", file.Path, len(file.Content))
	}

	if opts.DryRun {
		return nil
	}

	message := opts.CommitMessage
	if strings.TrimSpace(message) == "" {
		message = fmt.Sprintf("Sync Trivy SbomReports (%s)", time.Now().UTC().Format(time.RFC3339))
	}

	gh, err := githubsync.NewClient(opts.GitHubAPIURL, opts.GitHubToken)
	if err != nil {
		return err
	}
	result, err := gh.CommitFiles(ctx, githubsync.CommitRequest{
		Owner:         opts.GitHubOwner,
		Repo:          opts.GitHubRepo,
		Branch:        opts.GitHubBranch,
		Message:       message,
		Files:         files,
		DeleteMissing: opts.DeleteMissing,
		PathPrefix:    opts.PathPrefix,
	})
	if err != nil {
		return err
	}
	if result.Skipped {
		fmt.Fprintln(os.Stderr, "no changes; skipped commit")
		return nil
	}
	fmt.Fprintf(os.Stderr, "created commit %s\n", result.CommitSHA)
	if result.CommitURL != "" {
		fmt.Fprintf(os.Stderr, "%s\n", result.CommitURL)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
