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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WEATHER_REPORT_STATE_FAILED  = "Failed"
	WEATHER_REPORT_STATE_SUCCESS = "Succeess"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WeatherReportSpec defines the desired state of WeatherReport
type WeatherReportSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Location string `json:"location"`
	// +kubebuilder:default:=standard
	Units Unit `json:"units,omitempty"`
}

// WeatherReportStatus defines the observed state of WeatherReport
type WeatherReportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Unit             string `json:"unit"`
	Temperature      string `json:"temperature"`
	FeelsLike        string `json:"feels_like"`
	RelativeHumidity string `json:"relative_humidity"`
	Windspeed        string `json:"windspeed"`
	Winddirection    string `json:"winddirection"`
	State            string `json:"state"`
}

// +kubebuilder:validation:Enum=standard;metric;imperial
type Unit string

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Status the Report"
// +kubebuilder:printcolumn:name="Temperature",type="string",JSONPath=".status.temperature",description="Temperature"
// +kubebuilder:printcolumn:name="FeelsLike",type="string",JSONPath=".status.feels_like",description="Temperature"
// +kubebuilder:printcolumn:name="Humidity",type="string",JSONPath=".status.relative_humidity",description="Temperature"
// +kubebuilder:printcolumn:name="Windspeed",type="string",JSONPath=".status.windspeed",description="Temperature"
// +kubebuilder:printcolumn:name="Winddirection",type="string",JSONPath=".status.winddirection",description="Temperature"

// WeatherReport is the Schema for the weatherreports API
type WeatherReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WeatherReportSpec   `json:"spec,omitempty"`
	Status WeatherReportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WeatherReportList contains a list of WeatherReport
type WeatherReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WeatherReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WeatherReport{}, &WeatherReportList{})
}
