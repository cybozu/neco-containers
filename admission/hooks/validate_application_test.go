package hooks

import (
	"testing"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func fillApplication(name, project, repoURL string) *v1alpha1.Application {
	app := &v1alpha1.Application{}
	app.Name = name
	app.Namespace = "default"
	app.Spec.Project = project
	app.Spec.Source.RepoURL = repoURL
	return app
}

const (
	adminRepoURL  = "https://github.com/cybozu/admin-apps.git"
	tenantRepoURL = "https://github.com/cybozu/tenant-apps.git"
)

var applicationValidatorConfig = &ArgoCDApplicationValidatorConfig{
	[]ArgoCDApplicationRule{
		{adminRepoURL, []string{"default", "system"}},
	},
}

var _ = Describe("validate Application WebHook with ", func() {
	It("should allow admin App on admin repo", func() {
		err := k8sClient.Create(testCtx, fillApplication("test1", "admin", adminRepoURL))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny admin App on tenant repo", func() {
		err := k8sClient.Create(testCtx, fillApplication("test2", "default", tenantRepoURL))
		Expect(err).To(HaveOccurred())
	})

	It("should deny updating App with invalid repoURL", func() {
		app := fillApplication("test3", "admin", adminRepoURL)
		err := k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		app.Spec.Source.RepoURL = tenantRepoURL
		err = k8sClient.Update(testCtx, app)
		Expect(err).To(HaveOccurred())
	})
})

func TestIgnoreGitSuffix(t *testing.T) {
	v := argocdApplicationValidator{}
	withoutSuffix := "https://github.com/cybozu/admin-apps"
	if v.ignoreGitSuffix(adminRepoURL) != withoutSuffix {
		t.Errorf(`v.ignoreGitSuffix(adminRepoURL) != "https://github.com/cybozu/admin-apps". %q`, v.ignoreGitSuffix(adminRepoURL))
	}
	if v.ignoreGitSuffix(withoutSuffix) != withoutSuffix {
		t.Errorf(`v.ignoreGitSuffix(withoutSuffix) != "https://github.com/cybozu/admin-apps". %q`, v.ignoreGitSuffix(withoutSuffix))
	}
}
