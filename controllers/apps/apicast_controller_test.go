//go:build integration

package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
)

const testAPIcastEmbeddedConfigurationSecretName = "apicast-embedded-configuration"

var _ = Describe("APIcast controller", func() {
	var testNamespace string

	BeforeEach(CreateNamespaceCallback(&testNamespace))
	AfterEach(DeleteNamespaceCallback(&testNamespace))

	Context("Run directly without existing APIcast", func() {
		It("Should create successfully", func() {
			Expect(1).To(Equal(1))
		})
	})

	// Test basic APIcast deployment
	Context("Run with basic APIcast deployment", func() {
		It("Should create successfully", func() {
			const (
				retryInterval = time.Second * 5
			)

			start := time.Now()

			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicastName := "example-apicast"
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
				},
			}

			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the correspondig APIcast K8s Deployment has been created
			apicastDeploymentName := "apicast-" + apicastName
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := testClient().Get(context.Background(), apicastDeploymentLookupKey, createdDeployment)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "APIcast creation and availability took '%s'\n", elapsed)
		})
	})

	// Test APIcast deployment with ExposedHost

	Context("Run with APIcast with ExposedHost Deployment", func() {
		It("Should create successfully", func() {
			const (
				retryInterval = time.Second * 5
			)

			start := time.Now()

			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicastName := "example-apicast"
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					ExposedHost: &appsv1alpha1.APIcastExposedHost{
						Host: "apicast.example.com",
					},
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
				},
			}
			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the correspondig APIcast K8s Deployment has been created
			apicastDeploymentName := "apicast-" + apicastName
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := testClient().Get(context.Background(), apicastDeploymentLookupKey, createdDeployment)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			// Check that the correspondig IU K8s Ingress has been created
			apicastIngressName := "apicast-" + apicastName
			apicastIngressLookupKey := types.NamespacedName{Name: apicastIngressName, Namespace: testNamespace}
			createdIngress := &networkingv1.Ingress{}
			Eventually(func() bool {
				err := testClient().Get(context.Background(), apicastIngressLookupKey, createdIngress)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			elapsed := time.Since(start)
			By(fmt.Sprintf("APIcast creation and availability took %s seconds", elapsed))
		})
	})
})

func testAPIcastEmbeddedConfigurationContent() string {
	return `{
  "services": [
    {
      "proxy": {
        "policy_chain": [
          { "name": "apicast.policy.upstream",
            "configuration": {
              "rules": [{
                "regex": "/",
                "url": "http://echo-api.3scale.net"
              }]
            }
          }
        ]
      }
    }
  ]
}`
}

func testCreateAPIcastEmbeddedConfigurationSecret(ctx context.Context, namespace string) error {
	embeddedConfigurationContent := testAPIcastEmbeddedConfigurationContent()
	embeddedConfigSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAPIcastEmbeddedConfigurationSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"config.json": embeddedConfigurationContent,
		},
	}

	return testClient().Create(ctx, &embeddedConfigSecret)
}
