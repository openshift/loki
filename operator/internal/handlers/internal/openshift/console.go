package openshift

import (
	"context"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/loki/operator/internal/external/k8s"
)

const consoleName = "cluster"

// GetConsoleURL returns the OpenShift Console URL from the cluster Console resource.
func GetConsoleURL(ctx context.Context, k k8s.Client) (string, error) {
	key := client.ObjectKey{Name: consoleName}
	c := &configv1.Console{}
	if err := k.Get(ctx, key, c); err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimRight(c.Status.ConsoleURL, "/"), nil
}
