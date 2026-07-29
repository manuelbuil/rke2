package criticalconfig

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/rke2/tests/docker"
)

// This suite verifies the control-plane critical-config validation: a server
// joining an existing cluster with a divergent RKE2-specific critical setting
// (here, a different ingress-controller) must fail fast during bootstrap instead
// of being admitted and causing non-deterministic cluster behavior.

var ci = flag.Bool("ci", false, "running on CI")

var tc *docker.TestConfig

func Test_DockerCriticalConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	flag.Parse()
	RunSpecs(t, "Critical Config Validation Test Suite")
}

var _ = Describe("Critical config validation", Ordered, func() {
	Context("A server with divergent critical config", func() {
		It("brings up the first server with ingress-controller=traefik", func() {
			var err error
			tc, err = docker.NewTestConfig()
			Expect(err).NotTo(HaveOccurred())

			tc.ServerYaml = "ingress-controller: traefik\n"
			Expect(tc.ProvisionServers(1)).To(Succeed())
			Expect(docker.RestartCluster(tc.Servers[:1])).To(Succeed())

			// Wait until the first server's supervisor is serving its config so a
			// joining server can fetch it for comparison.
			Eventually(func() (string, error) {
				return tc.Servers[0].RunCmdOnNode("kubectl get --raw='/readyz' --kubeconfig=/etc/rancher/rke2/rke2.yaml")
			}, "180s", "5s").Should(ContainSubstring("ok"))
		})

		It("fails to join a second server with ingress-controller=ingress-nginx", func() {
			// The second server joins the first (RKE2_URL is set automatically for
			// non-zero server indexes) but declares a different ingress-controller.
			tc.ServerYaml = "ingress-controller: ingress-nginx\n"
			Expect(tc.ProvisionServers(2)).To(Succeed())

			// Start it without asserting success: bootstrap validation is expected
			// to abort the join.
			_, _ = tc.Servers[1].RunCmdOnNode("systemctl start rke2-server")

			Eventually(func() (string, error) {
				return tc.Servers[1].DumpServiceLogs(250)
			}, "180s", "5s").Should(
				ContainSubstring("critical configuration value mismatch between servers"),
				"second server should have failed bootstrap due to critical config mismatch",
			)

			// The offending field should be named to help the user.
			logs, err := tc.Servers[1].DumpServiceLogs(250)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainSubstring("critical-extra-config"))
		})
	})
})

var failed bool
var _ = AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var _ = AfterSuite(func() {
	if tc != nil && failed {
		AddReportEntry("journald-logs", tc.DumpServiceLogs(250))
	}
	if *ci || (tc != nil && !failed) {
		Expect(tc.Cleanup()).To(Succeed())
	}
})
