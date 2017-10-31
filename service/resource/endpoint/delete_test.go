package endpoint

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Endpoint_GetDeleteState(t *testing.T) {
	testCases := []struct {
		CurrentState        interface{}
		DesiredState        interface{}
		ExpectedDeleteState interface{}
		Obj                 interface{}
	}{
		{
			CurrentState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedDeleteState: Endpoint{
				IPs:              []string{},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			CurrentState: Endpoint{
				IPs: []string{
					"1.1.1.1",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedDeleteState: Endpoint{
				IPs: []string{
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			CurrentState: Endpoint{
				IPs: []string{
					"5.5.5.5",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedDeleteState: Endpoint{
				IPs: []string{
					"5.5.5.5",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			CurrentState: Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedDeleteState: Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			CurrentState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: nil,
			ExpectedDeleteState: Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			CurrentState:        nil,
			DesiredState:        nil,
			ExpectedDeleteState: Endpoint{},
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
		result, err := newResource.GetDeleteState(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedDeleteState, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedDeleteState, result)
		}
	}
}

func Test_Resource_Endpoint_ProcessDeleteState(t *testing.T) {
	testCases := []struct {
		DeleteState       interface{}
		ExpectedEndpoints []*apiv1.Endpoints
		SetupEndpoints    []*apiv1.Endpoints
		SetupService      *apiv1.Service
		Obj               interface{}
	}{
		{
			DeleteState: Endpoint{
				IPs: []string{
					"1.2.3.4",
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},

			ExpectedEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
								{
									IP: "1.1.1.1",
								},
							},
						},
					},
				},
			},
			SetupEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
							},
						},
					},
				},
			},
			SetupService: &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: apiv1.ServiceSpec{
					Ports: []apiv1.ServicePort{
						{
							Port: 1234,
						},
					},
				},
			},
		},
		{
			DeleteState: Endpoint{
				IPs: []string{
					"1.2.3.4",
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
								{
									IP: "1.1.1.1",
								},
							},
						},
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
								{
									IP: "1.1.1.1",
								},
							},
						},
					},
				},
			},
			SetupEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
							},
						},
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
							},
						},
					},
				},
			},
			SetupService: &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: apiv1.ServiceSpec{
					Ports: []apiv1.ServicePort{
						{
							Port: 1234,
						},
						{
							Port: 5678,
						},
					},
				},
			},
		},
		{
			DeleteState: Endpoint{
				IPs: []string{
					"1.2.3.4",
					"5.6.7.8",
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
								{
									IP: "5.6.7.8",
								},
								{
									IP: "1.1.1.1",
								},
							},
							Ports: []apiv1.EndpointPort{
								{
									Port: 1234,
								},
							},
						},
					},
				},
			},
			SetupEndpoints: []*apiv1.Endpoints{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []apiv1.EndpointSubset{
						{
							Addresses: []apiv1.EndpointAddress{
								{
									IP: "1.2.3.4",
								},
								{
									IP: "5.6.7.8",
								},
							},
							Ports: []apiv1.EndpointPort{
								{
									Port: 1234,
								},
							},
						},
					},
				},
			},
			SetupService: &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: apiv1.ServiceSpec{
					Ports: []apiv1.ServicePort{
						{
							Port: 1234,
						},
					},
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
		if tc.SetupService != nil {
			if _, err := fakeK8sClient.CoreV1().Services(tc.SetupService.Namespace).Create(tc.SetupService); err != nil {
				t.Fatalf("%d: error returned setting up k8s service: %s\n", i, err)
			}
		}
		for _, k8sEndpoint := range tc.SetupEndpoints {
			if _, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint); err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoints: %s\n", i, err)
			}
		}
		err := newResource.ProcessDeleteState(canceledcontext.NewContext(context.TODO(), make(chan struct{})), tc.Obj, tc.DeleteState)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		for _, k8sEndpoint := range tc.ExpectedEndpoints {
			returnedEndpoint, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Get(k8sEndpoint.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoints: %s\n", i, err)
			}
			if !reflect.DeepEqual(k8sEndpoint, returnedEndpoint) {
				t.Fatalf("case %d expected %#v got %#v", i+1, k8sEndpoint, returnedEndpoint)
			}
		}
	}
}
