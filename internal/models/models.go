package models

import "encoding/json"

type Record struct {
	ID            int        `json:"id" yaml:"id"`
	Status        string     `json:"status" yaml:"status"` // ON, OFF
	PollingPeriod int        `json:"polling_period" yaml:"polling_period"`
	Temperature   int        `json:"temperature"` // int type to avoid issues with float values
	LampStatus    string     `json:"lamp_status"` // ON, OFF
	Voltage       int        `json:"voltage"`     // same reason as Temperature
	Thresholds    Thresholds `json:"thresholds"`
}

type Thresholds struct {
	Temperature []SensorData `json:"temperature"`
	Humidity    []SensorData `json:"humidity"`
	Voltage     []SensorData `json:"voltage"`
}

type SensorData struct {
	Value         int `json:"value"`
	PollingPeriod int `json:"polling_period"`
}

func (t *Thresholds) UnmarshalJSON(data []byte) error {
	var rawThresholds map[string][]SensorData

	if err := json.Unmarshal(data, &rawThresholds); err != nil {
		return err
	}

	if temp, ok := rawThresholds["Temperature"]; ok {
		t.Temperature = temp
	}
	if hum, ok := rawThresholds["Humidity"]; ok {
		t.Humidity = hum
	}
	if volt, ok := rawThresholds["voltage"]; ok {
		t.Voltage = volt
	}

	return nil
}
