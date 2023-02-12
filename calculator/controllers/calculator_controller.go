/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	appv1alpha1 "git.epam.com/oleksandr_aloshchenko/edp/calculator/api/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CalculatorReconciler reconciles a Calculator object.
type CalculatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.calculator.com,resources=calculators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.calculator.com,resources=calculators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.calculator.com,resources=calculators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *CalculatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	//Fetch the Calculator instance
	calculator := &appv1alpha1.Calculator{}
	err := r.Get(ctx, req.NamespacedName, calculator)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("calculator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get calculator")
		return ctrl.Result{}, err
	}

	result := calculate(calculator.Spec.Xfield, calculator.Spec.Yfield, calculator.Spec.Operator)

	updateCalculatorStatus(calculator, result)

	//Update Calculator status
	err = r.Status().Update(ctx, calculator)
	if err != nil {
		logger.Error(err, "failed to update calculator status")
		return ctrl.Result{}, err
	}

	secret := createSecret(calculator)

	//Create secret
	err = r.Create(ctx, secret)
	if err != nil {
		logger.Error(err, "failed to create secret name", "secret name", secret.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CalculatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Calculator{}).
		Complete(r)
}

func calculate(x, y int, operator string) int {
	switch operator {
	case "+":
		return x + y
	case "-":
		return x - y
	case "*":
		return x * y
	default:
		return 0
	}
}

func updateCalculatorStatus(calc *appv1alpha1.Calculator, result int) {
	calc.Status.Processed = true
	calc.Status.Result = result
}

// createSecret returns a secret  with the same name of the CR
func createSecret(calc *appv1alpha1.Calculator) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      calc.Name,
			Namespace: calc.Namespace,
			Annotations: map[string]string{
				"managed-by": "calc-operator",
			},
		},
		Data: map[string][]byte{
			"result": []byte(strconv.Itoa(calc.Status.Result)),
		},
	}
	return secret
}
