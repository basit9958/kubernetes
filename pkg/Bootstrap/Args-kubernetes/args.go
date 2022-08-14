package flags

func Apiserverflags() []string {
	args := []string{
		"--advertise-address=10.0.0.1",
		"--allow-privileged=true",
		"--audit-log-maxage=30",
		"--audit-log-maxbackup=3",
		"--audit-log-maxsize=100",
		"--audit-log-path=/var/log/audit.log",
		"--authorization-mode=RBAC",
		"--bind-address=0.0.0.0",
		"--client-ca-file=ca.pem",
		"--enable-admission-plugins=NamespaceLifecycle,NodeRestriction,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota",
		"--etcd-cafile=/var/lib/kubernetes/ca.pem",
		"--etcd-certfile=/var/lib/kubernetes/kubernetes.pem",
		"--etcd-keyfile=/var/lib/kubernetes/kubernetes-key.pem",
		"--etcd-servers=",
		"--event-ttl=1h",
		"--encryption-provider-config=/var/lib/kubernetes/encryption-config.yaml",
		"--kubelet-certificate-authority=/var/lib/kubernetes/ca.pem",
		"--kubelet-client-certificate=/var/lib/kubernetes/kubernetes.pem",
		"--kubelet-client-key=/var/lib/kubernetes/kubernetes-key.pem",
		"--runtime-config='api/all=true'",
		"--service-account-key-file=/var/lib/kubernetes/service-account.pem",
		"--service-account-signing-key-file=/var/lib/kubernetes/service-account-key.pem",
		"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
		"--service-cluster-ip-range=10.32.0.0/24",
		"--service-node-port-range=30000-32767",
		"--token-auth-file=/var/lib/kubernetes/token.kubeconfig",
		"--tls-cert-file=/var/lib/kubernetes/kubernetes.pem",
		"--tls-private-key-file=/var/lib/kubernetes/kubernetes-key.pem",
		"--v=2",
	}
	return args
}

func Kubecontrollermanagerflags() []string {
	args := []string{
		"--authentication-kubeconfig=/var/lib/kubernetes/auth.kubeconfig",
		"--authorization-always-allow-paths=[/healthz]",
		"--bind-address=0.0.0.0",
		"--cluster-cidr=10.200.0.0/16",
		"--cluster-name=chaos-mesh",
		"--cluster-signing-cert-file=/var/lib/kubernetes/ca.pem",
		"--cluster-signing-key-file=/var/lib/kubernetes/ca-key.pem",
		"--kubeconfig=/var/lib/kubernetes/kube-controller-manager.kubeconfig",
		"--leader-elect=true",
		"--root-ca-file=ca.pem",
		"--service-account-private-key-file=/var/lib/kubernetes/service-account-key.pem",
		"--service-cluster-ip-range=10.32.0.0/24",
		"--use-service-account-credentials=true",
		"--v=2",
	}
	return args
}
