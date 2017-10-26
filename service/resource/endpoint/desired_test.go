package endpoint

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Endpoint_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		SetupEndpoints    []*v1.Endpoints
		ExpectedEndpoints []Endpoint
	}{
		{
			Obj: &Endpoint{
				IP:               "1.1.1.1",
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedEndpoints: []Endpoint{
				Endpoint{
					IP:               "1.1.1.1",
					ServiceName:      "TestService",
					ServiceNamespace: "TestNamespace",
				},
			},
		},
		{
			Obj: &Endpoint{
				IP:               "1.1.1.1",
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			SetupEndpoints: []*v1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{
								{
									IP: "1.1.1.1",
								},
							},
						},
					},
				},
			},
			ExpectedEndpoints: []Endpoint{
				Endpoint{
					IP:               "1.1.1.1",
					ServiceName:      "TestService",
					ServiceNamespace: "TestNamespace",
				},
			},
		},
		{
			Obj: &Endpoint{
				IP:               "1.1.1.1",
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			SetupEndpoints: []*v1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{
								{
									IP: "1.1.1.1",
								},
								{
									IP: "1.2.3.4",
								},
							},
						},
					},
				},
			},
			ExpectedEndpoints: []Endpoint{
				{
					IP:               "1.1.1.1",
					ServiceName:      "TestService",
					ServiceNamespace: "TestNamespace",
				},
				{
					IP:               "1.2.3.4",
					ServiceName:      "TestService",
					ServiceNamespace: "TestNamespace",
				},
			},
		},
	}
	var err error

	for i, tc := range testCases {
		fakeK8sClient := fake.NewSimpleClientset()
		var newResource *Resource
		{
			resourceConfig := DefaultConfig()
			resourceConfig.K8sClient = fakeK8sClient
			resourceConfig.Logger = microloggertest.New()
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
		}

		for _, k8sEndpoint := range tc.SetupEndpoints {
			if _, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint); err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoint: %s\n", i, err)
			}
		}
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedEndpoints, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedEndpoints, result)
		}
	}
}
