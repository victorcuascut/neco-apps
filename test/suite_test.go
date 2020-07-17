package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cybozu-go/log"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	if os.Getenv("SSH_PRIVKEY") == "" {
		t.Skip("no SSH_PRIVKEY envvar")
	}

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("/tmp/junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Test", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(30 * time.Minute)

	prepare()

	log.DefaultLogger().SetOutput(GinkgoWriter)

	fmt.Println("Begin tests...")
})

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("Test applications", func() {
	Context("prepareNodes", prepareNodes)
	if !doUpgrade && !doBootstrap {
		Context("prepareLoadPods", prepareLoadPods)
	}
	Context("setup", testSetup)
	if doBootstrap {
		return
	}
	if !doUpgrade {
		Context("OSDPodsSpread", testOSDPodsSpreadAll)
	}
	if doReboot {
		Context("reboot", testRebootAllNodes)
	}
	Context("network-policy", testNetworkPolicy)
	Context("metallb", testMetalLB)
	Context("external-dns", testExternalDNS)
	Context("cert-manager", testCertManager)
	Context("contour", testContour)
	Context("machines-endpoints", testMachinesEndpoints)
	Context("kube-state-metrics", testKubeStateMetrics)
	Context("prometheus", testPrometheus)
	Context("grafana-operator", testGrafanaOperator)
	Context("sandbox-grafana", testSandboxGrafana)
	Context("alertmanager", testAlertmanager)
	Context("pushgateway", testPushgateway)
	Context("ingress-health", testIngressHealth)
	Context("prometheus-metrics", testPrometheusMetrics)
	Context("metrics-server", testMetricsServer)
	Context("teleport", testTeleport)
	Context("topolvm", testTopoLVM)
	Context("elastic", testElastic)
	Context("argocd-ingress", testArgoCDIngress)
	Context("admission", testAdmission)
	Context("bmc-reverse-proxy", testBMCReverseProxy)
	Context("local-pv-provisioner", testLocalPVProvisioner)
	Context("rookOperator", testRookOperator)
	Context("MONPodsSpread", testMONPodsSpreadAll)
	Context("rookRGW", testRookRGW)
	Context("rookRBD", testRookRBDAll)
})
