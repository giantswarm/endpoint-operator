package endpoint

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	canceledcontext "github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Endpoint_GetCurrentState(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		ExpectedEndpoint interface{}
	}{
		{
			Obj: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip":      "1.1.1.1",
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			ExpectedEndpoint: &Endpoint{
				IP:               "1.1.1.1",
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			Obj: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			ExpectedEndpoint: nil,
		},
		{
			Obj: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip": "1.1.1.1",
					},
				},
			},
			ExpectedEndpoint: nil,
		},
		{
			Obj: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"jabber": "1.1.1.1",
						"wocky":  "TestService",
					},
				},
			},
			ExpectedEndpoint: nil,
		},
		{
			Obj: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
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
		result, err := newResource.GetCurrentState(canceledcontext.NewContext(context.TODO(), make(chan struct{})), tc.Obj)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedEndpoint, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedEndpoint, result)
		}
	}
}
