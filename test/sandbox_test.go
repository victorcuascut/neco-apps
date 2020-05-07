package test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testSandboxGrafana() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=sandbox",
				"get", "statefulset/grafana", "-o=json")
			if err != nil {
				return err
			}
			statefulSet := new(appsv1.StatefulSet)
			err = json.Unmarshal(stdout, statefulSet)
			if err != nil {
				return err
			}

			if int(statefulSet.Status.ReadyReplicas) != 1 {
				return fmt.Errorf("ReadyReplicas is not 1: %d", int(statefulSet.Status.ReadyReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should have data sources and dashboards", func() {
		By("getting external IP of grafana service")
		stdout, stderr, err := ExecAt(boot0, "kubectl", "--namespace=sandbox", "get", "services", "grafana", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		service := new(corev1.Service)
		err = json.Unmarshal(stdout, service)
		Expect(err).NotTo(HaveOccurred())
		loadBalancerIP := service.Status.LoadBalancer.Ingress[0].IP

		By("getting admin stats from grafana")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "curl", "-u", "admin:AUJUl1K2xgeqwMdZ3XlEFc1QhgEQItODMNzJwQme", loadBalancerIP+":3000/api/admin/stats")
			if err != nil {
				return fmt.Errorf("unable to get admin stats, stderr: %s, err: %v", stderr, err)
			}
			var adminStats struct {
				Dashboards  int `json:"dashboards"`
				Datasources int `json:"datasources"`
			}
			err = json.Unmarshal(stdout, &adminStats)
			if err != nil {
				return err
			}
			if adminStats.Datasources == 0 {
				return fmt.Errorf("no data sources")
			}
			if adminStats.Dashboards != 0 {
				return fmt.Errorf("%d dashboards exist", adminStats.Dashboards)
			}
			return nil
		}).Should(Succeed())
	})
}
