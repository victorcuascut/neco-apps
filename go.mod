module github.com/cybozu-go/neco-apps

go 1.13

replace (
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190805141520-2fe0317bcee0
	launchpad.net/gocheck => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
)

require (
	github.com/argoproj/argo-cd v1.1.0-rc7
	github.com/argoproj/pkg v0.0.0-20190708182346-fb13aebbef1c // indirect
	github.com/creack/pty v1.1.7
	github.com/cybozu-go/log v1.5.0
	github.com/cybozu-go/sabakan/v2 v2.4.2
	github.com/google/go-cmp v0.3.1
	github.com/jetstack/cert-manager v0.14.3
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/projectcontour/contour v1.0.1
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/common v0.7.0
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898 // indirect
	google.golang.org/appengine v1.6.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.11.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.5
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v0.17.5 // indirect
	sigs.k8s.io/yaml v1.1.0
)
