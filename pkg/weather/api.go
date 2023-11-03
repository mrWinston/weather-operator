package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrWinston/weather-operator/pkg/slicefuncs"
)

// how to call it?
// reconciler:
// input struct
// output struct based on api
// api wrapper:
// weather.get(input) --> returns output

type WeatherSeries struct {
	Time   []string
	Values map[string][]float64
}

func (w *WeatherSeries) UnmarshalJSON(data []byte) error {

	// first, unmarshal the object into a map of raw json
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// next, unmarshal the status and message fields, and any
	// other fields that don't belong to the "CoolKey" group
	if err := json.Unmarshal(m["time"], &w.Time); err != nil {
		var timeSolo string
		if err := json.Unmarshal(m["time"], &timeSolo); err != nil {
			return err
		}
		w.Time = append(w.Time, timeSolo)
	}
	delete(m, "time")

	w.Values = make(map[string][]float64)
	for k, v := range m {
		var currentSeries []float64
		if err := json.Unmarshal(v, &currentSeries); err != nil {
			var currentSolo float64
			if err := json.Unmarshal(v, &currentSolo); err != nil {
				return err
			}
			currentSeries = append(currentSeries, currentSolo)
		}
		w.Values[k] = currentSeries
	}

	return nil
}

type WeatherOutput struct {
	Latitude             float64           `json:"latitude,omitempty"`
	Longitude            float64           `json:"longitude,omitempty"`
	Elevation            float64           `json:"elevation,omitempty"`
	GenerationtimeMs     float64           `json:"generationtime_ms,omitempty"`
	UTCOffsetSeconds     int               `json:"utc_offset_seconds,omitempty"`
	Timezone             string            `json:"timezone,omitempty"`
	TimezoneAbbreviation string            `json:"timezone_abbreviation,omitempty"`
	Hourly               WeatherSeries     `json:"hourly,omitempty"`
	HourlyUnits          map[string]string `json:"hourly_units,omitempty"`
	Daily                WeatherSeries     `json:"daily,omitempty"`
	DailyUnits           map[string]string `json:"daily_units,omitempty"`
	Minutely15           WeatherSeries     `json:"minutely_15,omitempty"`
	Minutely15Units      map[string]string `json:"minutely_15_units,omitempty"`
	Current              WeatherSeries     `json:"current,omitempty"`
	CurrentUnits         map[string]string `json:"current_units,omitempty"`
}

type WeatherInput struct {
	Latitude          float64
	Longitude         float64
	Elevation         float64
	Minutely15        []string
	Hourly            []string
	Daily             []string
	Current           []string
	TemperaturUnit    string
	WindspeedUnit     string
	PrecipitationUnit string
	Timeformat        string
	Timezone          string
	PastDays          int
	ForecastDays      int
	StartDate         string
	EndDate           string
	ApiKey            string
}

type InvalidInputError struct {
	Arg    string
	Val    interface{}
	Reason string
}

func (err *InvalidInputError) Error() string {
	return fmt.Sprintf("Argument '%s' invalid value: %v: %s", err.Arg, err.Val, err.Reason)
}

func NewInvalidInputerror(argument string, value interface{}, reason string) *InvalidInputError {
	return &InvalidInputError{
		Arg:    argument,
		Val:    value,
		Reason: reason,
	}
}

func (w *WeatherInput) Validate() []error {
	errs := []error{}
	// check required fields first
	if w.Latitude == 0 {
		errs = append(errs, NewInvalidInputerror("Latitude", nil, "Must be set"))
	}
	if w.Longitude == 0 {
		errs = append(errs, NewInvalidInputerror("Longitude", nil, "Must be set"))
	}
	if len(w.Daily)+len(w.Hourly)+len(w.Minutely15)+len(w.Current) == 0 {
		errs = append(errs, NewInvalidInputerror("Daily,Hourly,Minutely15,Current", nil, "At least one of them must be set"))
	}

	for _, v := range w.Daily {
		if slicefuncs.FindIndexFunc(WEATHER_VALID_DAILY_VARS, func(elem string) bool {
			return elem == v
		}) == -1 {
			errs = append(errs, NewInvalidInputerror("Daily", v, "Unsupported Variable"))
		}
	}
	for _, v := range w.Hourly {
		if slicefuncs.FindIndexFunc(WEATHER_VALID_HOURLY_VARS, func(elem string) bool {
			return elem == v
		}) == -1 {
			errs = append(errs, NewInvalidInputerror("Hourly", v, "Unsupported Variable"))
		}
	}
	for _, v := range w.Minutely15 {
		if slicefuncs.FindIndexFunc(WEATHER_VALID_MINUTELY_15_VARS, func(elem string) bool {
			return elem == v
		}) == -1 {
			errs = append(errs, NewInvalidInputerror("Minutely15", v, "Unsupported Variable"))
		}
	}
	for _, v := range w.Current {
		if slicefuncs.FindIndexFunc(WEATHER_VALID_CURRENT_VARS, func(elem string) bool {
			return elem == v
		}) == -1 {
			errs = append(errs, NewInvalidInputerror("Current", v, "Unsupported Variable"))
		}
	}

	return errs
}

func (w *WeatherInput) GetQueryString() (string, error) {
	errs := w.Validate()

	if len(errs) != 0 {
		var errMsg strings.Builder
		errMsg.WriteString("Multiple Validation errors: \n")
		for _, err := range errs {
			errMsg.WriteString("\t")
			errMsg.WriteString(err.Error())
			errMsg.WriteString("\n")
		}
		return "", errors.New(errMsg.String())
	}

	params := make(url.Values)

	params.Add(WEATHER_PARAMETER_LATITUDE, fmt.Sprintf("%g", w.Latitude))
	params.Add(WEATHER_PARAMETER_LONGITUDE, fmt.Sprintf("%g", w.Longitude))
	params.Add(WEATHER_PARAMETER_ELEVATION, fmt.Sprintf("%g", w.Elevation))
	if len(w.Daily) != 0 {
		params.Add(WEATHER_PARAMETER_DAILY, strings.Join(w.Daily, ","))
	}
	if len(w.Hourly) != 0 {
		params.Add(WEATHER_PARAMETER_HOURLY, strings.Join(w.Hourly, ","))
	}
	if len(w.Current) != 0 {
		params.Add(WEATHER_PARAMETER_CURRENT, strings.Join(w.Current, ","))
	}
	if len(w.Minutely15) != 0 {
		params.Add(WEATHER_PARAMETER_MINUTELY_15, strings.Join(w.Minutely15, ","))
	}

	if w.TemperaturUnit != "" {
		params.Add(WEATHER_PARAMETER_TEMPERATURE_UNIT, w.TemperaturUnit)
	}
	if w.WindspeedUnit != "" {
		params.Add(WEATHER_PARAMETER_WINDSPEED_UNIT, w.WindspeedUnit)
	}
	if w.PrecipitationUnit != "" {
		params.Add(WEATHER_PARAMETER_PRECIPITATION_UNIT, w.PrecipitationUnit)
	}
	if w.Timeformat != "" {
		params.Add(WEATHER_PARAMETER_TIMEFORMAT, w.Timeformat)
	}
	if w.Timezone != "" {
		params.Add(WEATHER_PARAMETER_TIMEZONE, w.Timezone)
	}
	params.Add(WEATHER_PARAMETER_PAST_DAYS, fmt.Sprintf("%d", w.PastDays))
	params.Add(WEATHER_PARAMETER_FORECAST_DAYS, fmt.Sprintf("%d", w.ForecastDays))
	if w.StartDate != "" {
		params.Add(WEATHER_PARAMETER_START_DATE, w.StartDate)
	}
	if w.EndDate != "" {
		params.Add(WEATHER_PARAMETER_END_DATE, w.EndDate)
	}

	if w.ApiKey != "" {
		params.Add(WEATHER_PARAMETER_APIKEY, w.ApiKey)
	}

	return params.Encode(), nil
}

type GeoLocation struct {
	Id          int      `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Latitude    float64  `json:"latitude,omitempty"`
	Longitude   float64  `json:"longitude,omitempty"`
	Elevation   float64  `json:"elevation,omitempty"`
	FeatureCode string   `json:"feature_code,omitempty"`
	CountryCode string   `json:"country_code,omitempty"`
	Admin1Id    int      `json:"admin1_id,omitempty"`
	Admin2Id    int      `json:"admin2_id,omitempty"`
	Admin3Id    int      `json:"admin3_id,omitempty"`
	Admin4Id    int      `json:"admin4_id,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	Population  int      `json:"population,omitempty"`
	Postcodes   []string `json:"postcodes,omitempty"`
	CountryId   int      `json:"country_id,omitempty"`
	Country     string   `json:"country,omitempty"`
	Admin1      string   `json:"admin1,omitempty"`
	Admin2      string   `json:"admin2,omitempty"`
	Admin3      string   `json:"admin3,omitempty"`
	Admin4      string   `json:"admin4,omitempty"`
}

type GeoLocationOutput struct {
	Results          []GeoLocation
	GenerationtimeMs float64
}

func GetWeatherReport(r *WeatherInput) (*WeatherOutput, error) {
	weatherEndpoint, err := url.Parse("https://api.open-meteo.com/v1/forecast")
	if err != nil {
		return nil, err
	}

	queryString, err := r.GetQueryString()
	if err != nil {
		return nil, err
	}

	weatherEndpoint.RawQuery = queryString

	client := &http.Client{}

	resp, err := client.Get(weatherEndpoint.String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wo WeatherOutput

	err = json.Unmarshal(body, &wo)

	return &wo, err
}

func NameToLocation(name string) (latitude float64, longitude float64, err error) {

	client := &http.Client{}
	geocordingEndpoint, err := url.Parse("https://geocoding-api.open-meteo.com/v1/search")
	if err != nil {
		return
	}
	queryParams := make(url.Values)
	queryParams.Add("format", "json")
	queryParams.Add("name", name)
	queryParams.Add("count", "1")
	geocordingEndpoint.RawQuery = queryParams.Encode()

	resp, err := client.Get(geocordingEndpoint.String())
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var location GeoLocationOutput
	if err = json.Unmarshal(body, &location); err != nil {
		return
	}

	if len(location.Results) == 0 {
		err = fmt.Errorf("No Location Found")
		return
	}

	return location.Results[0].Latitude, location.Results[0].Longitude, nil
}
