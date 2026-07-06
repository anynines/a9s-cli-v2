package klutch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/k8s"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	controlPlaneInfoConfigMapName      = "klutch-control-plane-info"
	controlPlaneInfoConfigMapNamespace = "default"
)

// LoadControlPlaneInfo reads the control plane info file written during install.
// Returns an error if the file is missing or invalid.
func LoadControlPlaneInfo(workDir string) (ControlPlaneClusterInfo, error) {
	filePath := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ControlPlaneClusterInfo{}, fmt.Errorf("could not read control-plane info file %s: %w", filePath, err)
	}
	var info ControlPlaneClusterInfo
	if err := yaml.Unmarshal(data, &info); err != nil {
		return ControlPlaneClusterInfo{}, fmt.Errorf("could not parse control-plane info file %s: %w", filePath, err)
	}
	return info, nil
}

// LoadControlPlaneInfoFromCluster reads the control-plane info from a ConfigMap in the control-plane cluster.
// kubeContext can be empty to use the current context.
func LoadControlPlaneInfoFromCluster(kubeContext string) (ControlPlaneClusterInfo, error) {
	kc := k8s.NewKubeClient(kubeContext)
	clientset := kc.GetKubernetesClientSet()
	cm, err := clientset.CoreV1().ConfigMaps(controlPlaneInfoConfigMapNamespace).Get(context.Background(), controlPlaneInfoConfigMapName, metav1.GetOptions{})
	if err != nil {
		return ControlPlaneClusterInfo{}, err
	}
	info := ControlPlaneClusterInfo{
		Host:        cm.Data["host"],
		IngressPort: cm.Data["ingressPort"],
	}
	return info, nil
}

// SaveControlPlaneInfoToCluster persists control-plane info in a ConfigMap for later discovery.
func SaveControlPlaneInfoToCluster(kubeContext string, info ControlPlaneClusterInfo) error {
	kc := k8s.NewKubeClient(kubeContext)
	clientset := kc.GetKubernetesClientSet()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneInfoConfigMapName,
			Namespace: controlPlaneInfoConfigMapNamespace,
		},
		Data: map[string]string{
			"host":        info.Host,
			"ingressPort": info.IngressPort,
		},
	}
	_, err := clientset.CoreV1().ConfigMaps(controlPlaneInfoConfigMapNamespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		_, err = clientset.CoreV1().ConfigMaps(controlPlaneInfoConfigMapNamespace).Create(context.Background(), cm, metav1.CreateOptions{})
	}
	return err
}

// DefaultBindURLFromInfo builds a bind URL using the stored host/port.
// It assumes https unless the port is explicitly 80.
func DefaultBindURLFromInfo(info ControlPlaneClusterInfo) string {
	host := info.Host
	port := info.IngressPort
	if host == "" {
		return ""
	}
	scheme := "https"
	if port == "80" {
		scheme = "http"
	}
	switch port {
	case "", "443", "80":
		return fmt.Sprintf("%s://%s/bind-noninteractive", scheme, host)
	default:
		return fmt.Sprintf("%s://%s:%s/bind-noninteractive", scheme, host, port)
	}
}
