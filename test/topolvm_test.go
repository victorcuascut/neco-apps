package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func testTopoLVM() {
	ns := "test-topolvm"
	It("should create test-topolvm namespace", func() {
		ExecSafeAt(boot0, "kubectl", "delete", "namespace", ns, "--ignore-not-found=true")
		ExecSafeAt(boot0, "kubectl", "create", "namespace", ns)
	})

	It("should be mounted in specified path", func() {
		By("deploying Pod with PVC")
		podYAML := `apiVersion: v1
kind: Pod
metadata:
  name: ubuntu
  labels:
    app.kubernetes.io/name: ubuntu
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu:18.04
    command: ["sleep", "infinity"]
    volumeMounts:
    - name: my-volume
      mountPath: /test1
  volumes:
  - name: my-volume
    persistentVolumeClaim:
      claimName: topo-pvc
`
		claimYAML := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: topo-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: topolvm-provisioner
`
		stdout, stderr, err := ExecAtWithInput(boot0, []byte(claimYAML), "kubectl", "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(podYAML), "kubectl", "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the specified volume exists in the Pod")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "pvc", "topo-pvc", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pods", "ubuntu", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "mountpoint", "-d", "/test1")
			if err != nil {
				return fmt.Errorf("failed to check mount point. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "grep", "/test1", "/proc/mounts")
			if err != nil {
				return err
			}
			fields := strings.Fields(string(stdout))
			if len(fields) < 3 {
				return errors.New("invalid mount information: " + string(stdout))
			}
			if fields[2] != "xfs" {
				return errors.New("/test1 is not xfs")
			}
			return nil
		}).Should(Succeed())

		By("writing file under /test1")
		writePath := "/test1/bootstrap.log"
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cp", "/etc/passwd", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "sync")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cat", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Expect(strings.TrimSpace(string(stdout))).ShouldNot(BeEmpty())

		By("getting node name where pod is placed")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "-n", ns, "get", "pods/ubuntu", "-o", "json")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		var pod corev1.Pod
		err = json.Unmarshal(stdout, &pod)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", stdout)
		nodeName := pod.Spec.NodeName

		By("rebooting the node")
		ExecSafeAt(boot0, "ckecli", "sabakan", "disable")
		stdout, stderr, err = ExecAt(boot0, "neco", "ipmipower", "restart", nodeName)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		time.Sleep(5 * time.Second)

		// By("confirming that the file survives")
		// skip temporarily due to ckecli or Kubernetes issue
		// Eventually(func() error {
		// 	stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cat", writePath)
		// 	if err != nil {
		// 		return fmt.Errorf("failed to cat. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		// 	}
		// 	if len(strings.TrimSpace(string(stdout))) == 0 {
		// 		return errors.New(writePath + " is empty")
		// 	}
		// 	return nil
		// }).Should(Succeed())

		By("confirming the pods on the rebooted node come back")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "pods", "--all-namespaces", "--field-selector", "spec.nodeName="+nodeName, "-o", "json")
			if err != nil {
				return fmt.Errorf("kubectl failed; stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			podList := new(corev1.PodList)
			err = json.Unmarshal(stdout, podList)
			if err != nil {
				return err
			}

			for _, pod := range podList.Items {
				if !(pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded) {
					return fmt.Errorf("Pod %s is %s", pod.Name, pod.Status.Phase)
				}
			}
			return nil
		}).Should(Succeed())
	})
}
