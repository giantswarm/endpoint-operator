package endpoint

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Endpoint_GetCurrentState(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		ExpectedEndpoint *Endpoint
	}{
		{
			Obj:              &apiv1.Pod{},
			ExpectedEndpoint: nil,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetCurrentState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedEndpoint, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedEndpoint, result)
		}
	}
}
