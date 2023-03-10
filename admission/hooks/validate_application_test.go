package hooks

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func fillApplication(name, project, repoURL string, repoURLs []string) (*unstructured.Unstructured, error) {
	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Kind: "Application", Version: "v1alpha1"})
	app.SetName(name)
	app.SetNamespace("default")
	err := unstructured.SetNestedField(app.UnstructuredContent(), project, "spec", "project")
	if err != nil {
		return nil, err
	}

	if len(repoURL) != 0 {
		err := unstructured.SetNestedField(app.UnstructuredContent(), repoURL, "spec", "source", "repoURL")
		if err != nil {
			return nil, err
		}
	}

	if len(repoURLs) != 0 {
		sources := make([]interface{}, len(repoURLs))
		for i, r := range repoURLs {
			sources[i] = map[string]interface{}{"repoURL": r}
		}
		err := unstructured.SetNestedSlice(app.UnstructuredContent(), sources, "spec", "sources")
		if err != nil {
			return nil, err
		}
	}

	// spec.destination is required
	err = unstructured.SetNestedMap(app.UnstructuredContent(), map[string]interface{}{}, "spec", "destination")
	if err != nil {
		return nil, err
	}

	return app, nil
}

const (
	adminRepoURL     = "https://github.com/cybozu/admin-apps.git"
	adminOrgURL      = "https://github.com/cybozu-admin"
	adminOrgRepoURL  = "https://github.com/cybozu-admin/admin-apps.git"
	tenantRepoURL    = "https://github.com/cybozu/tenant-apps.git"
	tenantOrgRepoURL = "https://github.com/cybozu-tenant/tenant-apps.git"
)

var applicationValidatorConfig = &ArgoCDApplicationValidatorConfig{
	[]ArgoCDApplicationRule{
		{adminRepoURL, "", []string{"default", "admin"}},
		{"", adminOrgURL, []string{"default", "admin"}},
	},
}

var _ = Describe("Application WebHook", func() {
	It("should allow admin App on admin repo", func() {
		app, err := fillApplication("test1", "admin", adminRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow admin App on admin organization", func() {
		app, err := fillApplication("test2", "admin", adminOrgRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny admin App on tenant repo", func() {
		app, err := fillApplication("test3", "admin", tenantRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny admin App on tenant organization", func() {
		app, err := fillApplication("test4", "admin", tenantOrgRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny updating App with invalid repoURL", func() {
		app, err := fillApplication("test5", "admin", adminRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		err = unstructured.SetNestedField(app.UnstructuredContent(), tenantRepoURL, "spec", "source", "repoURL")
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Update(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny updating App with invalid organization repoURL", func() {
		app, err := fillApplication("test6", "admin", adminOrgRepoURL, nil)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		err = unstructured.SetNestedField(app.UnstructuredContent(), tenantOrgRepoURL, "spec", "source", "repoURL")
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Update(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should allow admin repos for admin project", func() {
		app, err := fillApplication("admin-repos-admin-project", "admin", "", []string{adminRepoURL, adminOrgRepoURL})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny tenant repo in multiple sources for admin project", func() {
		app, err := fillApplication("tenant-repo-multiple-sources-admin-project", "admin", "", []string{adminRepoURL, tenantRepoURL})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny adding tenant repos for admin project", func() {
		app, err := fillApplication("add-tenant-repos-admin-project", "admin", adminRepoURL, []string{adminOrgRepoURL})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		sources := []interface{}{
			map[string]interface{}{"repoURL": adminOrgRepoURL},
			map[string]interface{}{"repoURL": tenantRepoURL},
			map[string]interface{}{"repoURL": tenantOrgRepoURL},
		}
		err = unstructured.SetNestedSlice(app.UnstructuredContent(), sources, "spec", "sources")
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Update(testCtx, app)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
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
