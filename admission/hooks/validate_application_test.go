package hooks

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func fillApplication(name, project, repoURL string) (*unstructured.Unstructured, error) {
	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Kind: "Application", Version: "v1alpha1"})
	app.SetName(name)
	app.SetNamespace("default")
	err := unstructured.SetNestedField(app.UnstructuredContent(), project, "spec", "project")
	if err != nil {
		return nil, err
	}
	err = unstructured.SetNestedField(app.UnstructuredContent(), repoURL, "spec", "source", "repoURL")
	if err != nil {
		return nil, err
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
		{"", adminOrgRepoURL, []string{"default", "admin"}},
	},
}

var _ = Describe("validate Application WebHook with ", func() {
	It("should allow admin App on admin repo", func() {
		app, err := fillApplication("test1", "admin", adminRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow admin App on admin organization", func() {
		app, err := fillApplication("test2", "admin", adminOrgRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny admin App on tenant repo", func() {
		app, err := fillApplication("test3", "admin", tenantRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).To(HaveOccurred())
	})

	It("should deny admin App on tenant organization", func() {
		app, err := fillApplication("test4", "admin", tenantOrgRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).To(HaveOccurred())
	})

	It("should deny updating App with invalid repoURL", func() {
		app, err := fillApplication("test5", "admin", adminRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		err = unstructured.SetNestedField(app.UnstructuredContent(), tenantRepoURL, "spec", "source", "repoURL")
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Update(testCtx, app)
		Expect(err).To(HaveOccurred())
	})

	It("should deny updating App with invalid organization repoURL", func() {
		app, err := fillApplication("test6", "admin", adminOrgRepoURL)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(testCtx, app)
		Expect(err).NotTo(HaveOccurred())

		err = unstructured.SetNestedField(app.UnstructuredContent(), tenantOrgRepoURL, "spec", "source", "repoURL")
		Expect(err).NotTo(HaveOccurred())
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
