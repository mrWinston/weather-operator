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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	weatherv1alpha1 "github.com/mrWinston/weather-operator/api/v1alpha1"
	"github.com/mrWinston/weather-operator/pkg/weather"
)

// WeatherReportReconciler reconciles a WeatherReport object
type WeatherReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=weather.mrwinston.github.io,resources=weatherreports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=weather.mrwinston.github.io,resources=weatherreports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=weather.mrwinston.github.io,resources=weatherreports/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WeatherReport object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *WeatherReportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// TODO(user): your logic here
	report := &weatherv1alpha1.WeatherReport{}
	err := r.Client.Get(ctx, req.NamespacedName, report)

	if err != nil {
		reqLogger.Error(err, "Can't retrieve WeatherReport CR")
		return ctrl.Result{}, err
	}

	lat, lon, err := weather.NameToLocation(report.Spec.Location)

	if err != nil {
		reqLogger.Error(err, "Can not find location")
		report.Status.State = weatherv1alpha1.WEATHER_REPORT_STATE_FAILED
		updErr := r.Client.Update(ctx, report)
		if updErr != nil {
			reqLogger.Error(updErr, "ErrorUpdating Status")
			return ctrl.Result{}, updErr
		}
		return ctrl.Result{}, err
	}

	weatherInput := &weather.WeatherInput{
		Latitude:  lat,
		Longitude: lon,
		Current: []string{
			weather.WEATHER_VAR_CURRENT_RELATIVEHUMIDITY_2M,
			weather.WEATHER_VAR_CURRENT_WINDDIRECTION_10M,
			weather.WEATHER_VAR_CURRENT_WINDSPEED_10M,
			weather.WEATHER_VAR_CURRENT_TEMPERATURE_2M,
			weather.WEATHER_VAR_CURRENT_APPARENT_TEMPERATURE,
		},
	}

	wo, err := weather.GetWeatherReport(weatherInput)
	if err != nil {
		reqLogger.Error(err, "Error getting Report")
		report.Status.State = weatherv1alpha1.WEATHER_REPORT_STATE_FAILED
		updErr := r.Status().Update(ctx, report)
		if updErr != nil {
			reqLogger.Error(updErr, "ErrorUpdating Status")
			return ctrl.Result{}, updErr
		}
		return ctrl.Result{}, err
	}

	reqLogger.Info("Got requested Report", "fullReport", wo)
	report.Status.Temperature = fmt.Sprintf("%g", wo.Current.Values[weather.WEATHER_VAR_CURRENT_TEMPERATURE_2M][0])
	report.Status.FeelsLike = fmt.Sprintf("%g", wo.Current.Values[weather.WEATHER_VAR_CURRENT_APPARENT_TEMPERATURE][0])
	report.Status.Windspeed = fmt.Sprintf("%g", wo.Current.Values[weather.WEATHER_VAR_CURRENT_WINDSPEED_10M][0])
	report.Status.Winddirection = fmt.Sprintf("%g", wo.Current.Values[weather.WEATHER_VAR_CURRENT_WINDDIRECTION_10M][0])
	report.Status.RelativeHumidity = fmt.Sprintf("%g", wo.Current.Values[weather.WEATHER_VAR_CURRENT_RELATIVEHUMIDITY_2M][0])
	report.Status.State = weatherv1alpha1.WEATHER_REPORT_STATE_SUCCESS
	updErr := r.Status().Update(ctx, report)
	if updErr != nil {
		reqLogger.Error(updErr, "ErrorUpdating Status")
		return ctrl.Result{}, updErr
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WeatherReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	rl := workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 10*time.Minute)
	return ctrl.NewControllerManagedBy(mgr).
		For(&weatherv1alpha1.WeatherReport{}).
		WithOptions(controller.Options{
			RateLimiter: rl,
		}).
		Complete(r)
}
