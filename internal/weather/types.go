package weather

import "time"

// Snapshot is the current weather at a location, in provider-neutral form.
type Snapshot struct {
	ObservedAt       time.Time
	FetchedAt        time.Time
	CloudCoverPct    int     // 0-100
	PrecipitationMMH float64 // rain, mm per hour
	SnowfallCMH      float64 // snow, cm per hour
	WindSpeedKMH     float64
	WindDirDeg       int // 0-359, direction the wind blows from
	Code             int // WMO weather code
	IsDay            bool
	TemperatureC     float64
}

// Condition is a coarse category used by the sky renderer.
type Condition string

const (
	ConditionUnknown Condition = "unknown"
	ConditionClear   Condition = "clear"
	ConditionCloudy  Condition = "cloudy"
	ConditionRain    Condition = "rain"
	ConditionSnow    Condition = "snow"
	ConditionFog     Condition = "fog"
	ConditionStorm   Condition = "storm"
)
