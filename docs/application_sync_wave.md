List of Application Sync Waves
====================================

The sync order of applications can be managed with the `argocd.argoproj.io/sync-wave` annotation.

| Application          | Wave |
| -------------------- | ---- |
| namespaces           | 1    |
| argocd               | 2    |
| local-pv-provisioner | 3    |
| secrets              | 3    |
| cert-manager         | 4    |
| external-dns         | 4    |
| metallb              | 4    |
| ingress              | 5    |
| topolvm              | 5    |
| unbound              | 5    |
| elastic              | 6    |
| rook                 | 6    |
| monitoring           | 7    |
| sandbox              | 7    |
| teleport             | 7    |
| network-policy       | 8    |
| argocd-ingress       | 9    |
| bmc-reverse-proxy    | 9    |
| metrics-server       | 9    |
| neco-admission       | 9    |
| team-management      | 9    |
| maneki-apps          | 10   |
