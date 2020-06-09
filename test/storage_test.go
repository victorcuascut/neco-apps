package test

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testLoadPods() {
	It("should deploy pods", func() {
		yamlCS := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: addload-for-cs
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: addload
  template:
    metadata:
      labels:
        app: addload
    spec:
      containers:
      - name: spread-test-ubuntu
        image: ubuntu:latest
        command: ["sleep", "infinity"]
        resources:
          requests:
            cpu: "2"
`
		stdout, stderr, err := ExecAtWithInput(boot0, []byte(yamlCS), "kubectl", "apply", "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		yamlSS := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: addload-for-ss
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: addload
  template:
    metadata:
      labels:
        app: addload
    spec:
      containers:
      - name: spread-test-ubuntu
        image: ubuntu:latest
        command: ["sleep", "infinity"]
        resources:
          requests:
            cpu: "1"
      nodeSelector:
        cke.cybozu.com/role: ss
      tolerations:
      - key: cke.cybozu.com/role
        operator: Equal
        value: storage
`
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(yamlSS), "kubectl", "apply", "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl",
				"get", "deployment", "addload-for-cs", "-o=json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return fmt.Errorf("stdout: %s, err: %v", stdout, err)
			}

			if deployment.Status.AvailableReplicas != 2 {
				return fmt.Errorf("addload-for-cs deployment's AvailableReplica is not 2: %d", int(deployment.Status.AvailableReplicas))
			}

			return nil
		}).Should(Succeed())

		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl",
				"get", "deployment", "addload-for-ss", "-o=json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return fmt.Errorf("stdout: %s, err: %v", stdout, err)
			}

			if deployment.Status.AvailableReplicas != 2 {
				return fmt.Errorf("addload-for-ss deployment's AvailableReplica is not 2: %d", int(deployment.Status.AvailableReplicas))
			}

			return nil
		}).Should(Succeed())
	})
}

func testRookOperator() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=rook-ceph",
				"get", "deployment/rook-ceph-operator", "-o=json")
			if err != nil {
				return err
			}

			ss := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, ss)
			if err != nil {
				return err
			}

			if ss.Status.AvailableReplicas != 1 {
				return fmt.Errorf("rook operator deployment's AvialbleReplicas is not 1: %d", int(ss.Status.ReadyReplicas))
			}
			return nil
		}).Should(Succeed())
	})
}

func testRookRGW() {
	It("should create test-rook-rgw namespace", func() {
		ExecSafeAt(boot0, "kubectl", "delete", "namespace", "test-rook-rgw", "--ignore-not-found=true")
		ExecSafeAt(boot0, "kubectl", "create", "namespace", "test-rook-rgw")
	})

	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=ceph-for-ss",
				"get", "cephcluster ceph-for-ss", "-o", "jsonpath='{.status.ceph.health}'")
			if err != nil {
				return err
			}
			health := strings.TrimSpace(string(stdout))
			if health != "HEALTH_OK" {
				return fmt.Errorf("ceph cluster is not HEALTH_OK: %s", health)
			}
			return nil
		}).Should(Succeed())
	})

	It("should deploy OSD POD successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=ceph-for-ss",
				"get", "deployment/rook-ceph-osd-0", "-o=json")
			if err != nil {
				return err
			}

			ss := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, ss)
			if err != nil {
				return err
			}

			if ss.Status.AvailableReplicas != *ss.Spec.Replicas {
				return fmt.Errorf("OSD deployment's ReadyReplica is not %d: %d", int(*ss.Spec.Replicas), int(ss.Status.ReadyReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should spread OSD PODs on ceph-for-ss", func() {
		stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "node", "-l", "node-role.kubernetes.io/ss=true", "-o=json")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		nodes := new(corev1.NodeList)
		err = json.Unmarshal(stdout, nodes)
		Expect(err).ShouldNot(HaveOccurred())

		nodeCounts := make(map[string]int)
		for _, node := range nodes.Items {
			nodeCounts[node.Name] = 0
		}

		stdout, stderr, err = ExecAt(boot0, "kubectl", "--namespace=ceph-for-ss",
			"get", "pod", "-l", "app=rook-ceph-osd", "-o=json")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		pods := new(corev1.PodList)
		err = json.Unmarshal(stdout, pods)
		Expect(err).ShouldNot(HaveOccurred())

		for _, pod := range pods.Items {
			nodeCounts[pod.Spec.NodeName]++
		}

		var min int = math.MaxInt32
		var max int
		for _, v := range nodeCounts {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		Expect(max-min).Should(BeNumerically("<=", 1), "nodeCounts=%v", nodeCounts)

		rackCounts := make(map[string]int)
		for _, node := range nodes.Items {
			rackCounts[node.Labels["topology.kubernetes.io/zone"]] += nodeCounts[node.Name]
		}

		min = math.MaxInt32
		max = 0
		for _, v := range rackCounts {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		Expect(max-min).Should(BeNumerically("<=", 1), "rackCounts=%v", rackCounts)
	})

	It("should be used from a POD with a s3 client", func() {
		ns := "test-rook-rgw"
		podPvcYaml := fmt.Sprintf(`apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: pod-ob
  namespace: %s
spec:
  generateBucketName: obc-poc
  storageClassName: rook-ceph-bucket
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-ob
  namespace: %s
spec:
  containers:
  - name: mycontainer
    image: quay.io/cybozu/ubuntu-debug:18.04
    imagePullPolicy: Always
    args:
    - infinity
    command:
    - sleep
    envFrom:
    - configMapRef:
        name: pod-ob
    - secretRef:
        name: pod-ob`, ns, ns)

		_, stderr, err := ExecAtWithInput(boot0, []byte(podPvcYaml), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "-n", ns, "exec", "pod-ob", "--", "date")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		stdout, stderr, err := ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-ob", "--", "sh", "-c", `"echo foobar > /tmp/foobar"`)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-ob", "--", "sh", "-c",
			`"s3cmd put /tmp/foobar --no-ssl --host=\${BUCKET_HOST} --host-bucket= s3://\${BUCKET_NAME}"`)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		stdout, _, _ = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-ob", "--", "sh", "-c",
			`"s3cmd ls s3://\${BUCKET_NAME} --no-ssl --host=\${BUCKET_HOST} --host-bucket= s3://\${BUCKET_NAME}"`)
		Expect(stdout).NotTo(BeEmpty())

		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-ob", "--", "sh", "-c",
			`"s3cmd get s3://\${BUCKET_NAME}/foobar /tmp/downloaded --no-ssl --host=\${BUCKET_HOST} --host-bucket="`)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-ob", "--", "cat", "/tmp/downloaded")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	})
}

func testRookRBD() {
	It("should create test-rook-rbd namespace", func() {
		ExecSafeAt(boot0, "kubectl", "delete", "namespace", "test-rook-rbd", "--ignore-not-found=true")
		ExecSafeAt(boot0, "kubectl", "create", "namespace", "test-rook-rbd")
	})

	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=ceph-for-cs",
				"get", "cephcluster ceph-for-cs", "-o", "jsonpath='{.status.ceph.health}'")
			if err != nil {
				return err
			}
			health := strings.TrimSpace(string(stdout))
			if health != "HEALTH_OK" {
				return fmt.Errorf("ceph cluster is not HEALTH_OK: %s", health)
			}
			return nil
		}).Should(Succeed())
	})

	It("should deploy OSD POD successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=ceph-for-cs",
				"get", "deployment/rook-ceph-osd-0", "-o=json")
			if err != nil {
				return err
			}

			ss := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, ss)
			if err != nil {
				return err
			}

			if ss.Status.AvailableReplicas != *ss.Spec.Replicas {
				return fmt.Errorf("OSD deployment's ReadyReplica is not %d: %d", int(*ss.Spec.Replicas), int(ss.Status.ReadyReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should spread OSD PODs on ceph-for-cs", func() {
		stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "node", "-l", "node-role.kubernetes.io/cs=true", "-o=json")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		nodes := new(corev1.NodeList)
		err = json.Unmarshal(stdout, nodes)
		Expect(err).ShouldNot(HaveOccurred())

		nodeCounts := make(map[string]int)
		for _, node := range nodes.Items {
			nodeCounts[node.Name] = 0
		}

		stdout, stderr, err = ExecAt(boot0, "kubectl", "--namespace=ceph-for-cs",
			"get", "pod", "-l", "app=rook-ceph-osd", "-o=json")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		pods := new(corev1.PodList)
		err = json.Unmarshal(stdout, pods)
		Expect(err).ShouldNot(HaveOccurred())

		for _, pod := range pods.Items {
			nodeCounts[pod.Spec.NodeName]++
		}

		var min int = math.MaxInt32
		var max int
		for _, v := range nodeCounts {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		Expect(max-min).Should(BeNumerically("<=", 1), "nodeCounts=%v", nodeCounts)

		rackCounts := make(map[string]int)
		for _, node := range nodes.Items {
			rackCounts[node.Labels["topology.kubernetes.io/zone"]] += nodeCounts[node.Name]
		}

		min = math.MaxInt32
		max = 0
		for _, v := range rackCounts {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		Expect(max-min).Should(BeNumerically("<=", 1), "rackCounts=%v", rackCounts)
	})

	It("should be mounted to a path specified on a POD", func() {
		ns := "test-rook-rbd"
		podPvcYaml := fmt.Sprintf(`kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pod-rbd
  namespace: %s
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: rook-ceph-block
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-rbd
  namespace: %s
  labels:
    app.kubernetes.io/name: pod-rbd
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu-debug:18.04
    imagePullPolicy: Always
    command: ["/usr/local/bin/pause"]
    volumeMounts:
    - mountPath: /test1
      name: rbd-volume
  volumes:
    - name: rbd-volume
      persistentVolumeClaim:
        claimName: pod-rbd`, ns, ns)

		_, stderr, err := ExecAtWithInput(boot0, []byte(podPvcYaml), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-rbd", "--", "mountpoint", "-d", "/test1")
			if err != nil {
				return fmt.Errorf("failed to check mount point. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		writePath := "/test1/test.txt"
		stdout, stderr, err := ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-rbd", "--", "cp", "/etc/passwd", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-rbd", "--", "sync")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "pod-rbd", "--", "cat", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	})
}
