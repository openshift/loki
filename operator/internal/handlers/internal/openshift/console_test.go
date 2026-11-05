package openshift

import (
	"context"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/loki/operator/internal/external/k8s/k8sfakes"
)

func TestGetConsoleURL_ReturnError_WhenOtherThanNotFound(t *testing.T) {
	k := &k8sfakes.FakeClient{}

	k.GetStub = func(_ context.Context, _ types.NamespacedName, _ client.Object, _ ...client.GetOption) error {
		return apierrors.NewBadRequest("bad request")
	}

	_, err := GetConsoleURL(context.TODO(), k)
	require.Error(t, err)
}

func TestGetConsoleURL_ReturnEmpty_WhenNotFound(t *testing.T) {
	k := &k8sfakes.FakeClient{}

	k.GetStub = func(_ context.Context, _ types.NamespacedName, _ client.Object, _ ...client.GetOption) error {
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	got, err := GetConsoleURL(context.TODO(), k)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestGetConsoleURL_ReturnURL_WhenConsoleExists(t *testing.T) {
	k := &k8sfakes.FakeClient{}

	k.GetStub = func(_ context.Context, name types.NamespacedName, out client.Object, _ ...client.GetOption) error {
		if name.Name == consoleName {
			k.SetClientObject(out, &configv1.Console{
				Status: configv1.ConsoleStatus{
					ConsoleURL: "https://console.apps.example.com/",
				},
			})
			return nil
		}
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	got, err := GetConsoleURL(context.TODO(), k)
	require.NoError(t, err)
	require.Equal(t, "https://console.apps.example.com", got)
}
