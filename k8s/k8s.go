package k8s

// A KubeClient can interact with a specific k8s cluster context, or the current context if
// KubeContext is empty.
type KubeClient struct {
	KubeContext string
}

// NewKubeClient returns a KubeClient for the given kube context. If the context is empty,
// it is ignored.
func NewKubeClient(kubeContext string) *KubeClient {
	return &KubeClient{
		KubeContext: kubeContext,
	}
}
