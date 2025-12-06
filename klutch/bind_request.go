package klutch

import (
	"encoding/json"
	"fmt"
)

// GroupResource is the minimal structure for a bind request API entry.
type GroupResource struct {
	Group    string `json:"group"`
	Resource string `json:"resource"`
}

// DefaultBindRequestJSON returns a bind request JSON that includes all known exported services.
// clusterID is typically the tenant name or workload identifier.
func DefaultBindRequestJSON(clusterID string) ([]byte, error) {
	req := struct {
		ClusterID string          `json:"clusterID"`
		Apis      []GroupResource `json:"apis"`
	}{
		ClusterID: clusterID,
		Apis: []GroupResource{
			{Group: "anynines.com", Resource: "postgresqlinstances"},
			{Group: "anynines.com", Resource: "servicebindings"},
			{Group: "anynines.com", Resource: "backups"},
			{Group: "anynines.com", Resource: "restores"},
		},
	}
	out, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default bind request: %w", err)
	}
	return out, nil
}
