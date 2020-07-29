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
	SetDefaultEventuallyTimeout(20 * time.Minute)

	prepare()

	log.DefaultLogger().SetOutput(GinkgoWriter)

	fmt.Println("Begin tests...")
})

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("Test applications", func() {
	if doCeph {
		Context("prepareNodes", prepareNodes)
		Context("prepareLoadPods", prepareLoadPods)
		Context("setup", testSetup)
		Context("OSDPodsSpread", testOSDPodsSpreadAll)
		Context("rookOperator", testRookOperator)
		Context("MONPodsSpread", testMONPodsSpreadAll)
		Context("rookRGW", testRookRGW)
		Context("rookRBD", testRookRBDAll)
		return
	}

	Context("prepareNodes", prepareNodes)
	Context("setup", testSetup)
	if doBootstrap {
		return
	}
	if doReboot {
		Context("reboot", testRebootAllNodes)
	}

	// preparing resources before test to make things faster
	Context("preparing argocd-ingress", prepareArgoCDIngress)
	Context("preparing contour", prepareContour)
	Context("preparing elastic", prepareElastic)
	Context("preparing local-pv-provisioner", prepareLocalPVProvisioner)
	Context("preparing metallb", prepareMetalLB)
	Context("preparing pushgateway", preparePushgateway)
	Context("preparing ingress-health", prepareIngressHealth)
	Context("preparing grafana-operator", prepareGrafanaOperator)
	Context("preparing topolvm", prepareTopoLVM)
	Context("preparing network-policy", prepareNetworkPolicy) // this must be the last preparation.

	// running tests
	Context("cephcluster", testLVTag)
})
