package pg

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

/*
Executes: kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
*/
func FindPrimaryPodOfServiceInstance(namespace, serviceInstanceName string) string {

	instanceLabel := fmt.Sprintf("%s=%s", A8sPGServiceInstanceNameLabelKey, serviceInstanceName)

	label := fmt.Sprintf("%s,%s,%s", A8sPGLabelPrimary, A8sPGServiceInstanceAPIGroupLabel, instanceLabel)

	//TODO Implement FindFirstPodByLabel in k8s/kubernetes_workload.go
	podName, err := k8s.FindFirstPodByLabel(namespace, label)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't find primary pod of service instance: %s/%s", namespace, serviceInstanceName))
	}

	return podName
}

