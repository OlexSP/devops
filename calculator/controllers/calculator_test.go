package controllers

import (
	"context"
	appv1alpha1 "git.epam.com/oleksandr_aloshchenko/edp/calculator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	"k8s.io/api/core/v1"
)

func TestCreateSecret(t *testing.T) {
	tests := []struct {
		name string
		calc *appv1alpha1.Calculator
		want *v1.Secret
	}{
		{
			name: "case 1: create a secret with correct data",
			calc: &appv1alpha1.Calculator{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-calc",
					Namespace: "default",
				},
				Status: appv1alpha1.CalculatorStatus{
					Result:    123,
					Processed: true,
				},
			},
			want: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-calc",
					Namespace: "default",
					Annotations: map[string]string{
						"managed-by": "calc-operator",
					},
				},
				Data: map[string][]byte{
					"result": []byte("123"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			secret := createSecret(test.calc)
			assert.Equal(t, test.want, secret)
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := map[string]struct {
		x        int
		y        int
		operator string
		expected int
	}{
		"TestAddition":        {x: 1, y: 2, operator: "+", expected: 3},
		"TestSubtraction":     {x: 2, y: 1, operator: "-", expected: 1},
		"TestMultiplication":  {x: 2, y: 3, operator: "*", expected: 6},
		"TestInvalidOperator": {x: 2, y: 3, operator: "/", expected: 0},
	}

	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			result := calculate(test.x, test.y, test.operator)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestUpdateCalculatorStatus(t *testing.T) {
	tests := []struct {
		name     string
		calc     *appv1alpha1.Calculator
		result   int
		expected *appv1alpha1.Calculator
	}{
		{
			name:   "Case 1: Update Calculator status",
			calc:   &appv1alpha1.Calculator{},
			result: 10,
			expected: &appv1alpha1.Calculator{
				Status: appv1alpha1.CalculatorStatus{
					Processed: true,
					Result:    10,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updateCalculatorStatus(test.calc, test.result)
			assert.Equal(t, test.calc, test.expected)
		})
	}
}

type mockClient struct{ mock.Mock }

func newMockClient() *mockClient { return &mockClient{} }

func TestReconcile(t *testing.T) {
	mockClient := newMockClient()

	r := CalculatorReconciler{}
	calculator := &appv1alpha1.Calculator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: appv1alpha1.CalculatorSpec{
			Xfield:   1,
			Yfield:   2,
			Operator: "+",
		},
		Status: appv1alpha1.CalculatorStatus{},
	}
	request := ctrl.Request{
		NamespacedName: client.ObjectKey{
			Name:      "test",
			Namespace: "default",
		},
	}

	mockClient.On("Get", mock.Anything, &request.NamespacedName, calculator).Return(nil)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	result, err := r.Reconcile(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	mockClient.AssertExpectations(t)
}
