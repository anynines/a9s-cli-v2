package creator

import "fmt"

// A simple abstraction to manage Kubernetes clusters allowing
// the creation of Kubernetes clusters using different automation backends.

type KubernetesClusterSpec struct {
	Name                 string
	NrOfNodes            int
	NodeMemory           string // e.g. '8gb' memory per node
	InfrastructureRegion string
}

func (s KubernetesClusterSpec) String() string {
	return fmt.Sprintf("Nr of nodes: %d, Node memory: %s, Infrastructure region: %s", s.NrOfNodes, s.NodeMemory, s.InfrastructureRegion)
}

/*
A simple interface to manage Kubernetes clusters. Implementations of this interface
will allow the creation of Kubernetes clusters using different automation backends.

Examples:
- Minikube
- Kind
- Eks

Milestone 1: Create and Delete. No modification. No day 2 lifecycle management.
*/
type KubernetesCreator interface {

	/*
		Creates a Kubernetes cluster
	*/
	Create(spec KubernetesClusterSpec, unattendedMode bool)

	/*
		Checks whether a Kubernetes cluster exists.
		Existence means that a Kubernetes cluster resource is available.
		It does verify whether the cluster is running.
	*/
	Exists(clustername string) bool

	/*
		Checks whether the Kubernetes cluster is running by checking whether
		the Kubernetes cluster API is available.
		Does not verfiy whether the cluster in healthy.
	*/
	Running(clustername string) bool
	Delete(clustername string, unattendedMode bool)

	/*
		Returns the context for a given clustername.
		This name may vary among implementations, e.g.
		a8s-demo vs. kind-a8s-demo.
	*/
	GetContext(clustername string) string
}
