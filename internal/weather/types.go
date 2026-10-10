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

// ConditionFromCode groups a WMO weather interpretation code (as documented by
// Open-Meteo: 0-3 sky, 45/48 fog, 51-67 and 80-82 rain, 71-77 and 85/86 snow,
// 95-99 thunderstorm).
func ConditionFromCode(code int) Condition {
	switch {
	case code == 0 || code == 1:
		return ConditionClear
	case code == 2 || code == 3:
		return ConditionCloudy
	case code == 45 || code == 48:
		return ConditionFog
	case (code >= 51 && code <= 67) || (code >= 80 && code <= 82):
		return ConditionRain
	case (code >= 71 && code <= 77) || code == 85 || code == 86:
		return ConditionSnow
	case code >= 95 && code <= 99:
		return ConditionStorm
	default:
		return ConditionUnknown
	}
}

// View is what the sky needs, derived from a Snapshot. The weather code fills
// gaps the numeric fields leave (a rain code with 0 mm still has to look like rain).
type View struct {
	CloudCoverPct    int
	PrecipitationMMH float64
	SnowfallCMH      float64
	WindSpeedKMH     float64
	WindDirDeg       int
	FogDensity       float64 // 0..1
	Thunder          bool
	IsDay            bool
	TemperatureC     float64
	Code             int
	ObservedAt       time.Time
	FetchedAt        time.Time
}

// View derives the sky inputs.
func (s Snapshot) View() View {
	v := View{
		CloudCoverPct:    clampInt(s.CloudCoverPct, 0, 100),
		PrecipitationMMH: nonNegative(s.PrecipitationMMH),
		SnowfallCMH:      nonNegative(s.SnowfallCMH),
		WindSpeedKMH:     nonNegative(s.WindSpeedKMH),
		WindDirDeg:       ((s.WindDirDeg % 360) + 360) % 360,
		IsDay:            s.IsDay,
		TemperatureC:     s.TemperatureC,
		Code:             s.Code,
		ObservedAt:       s.ObservedAt,
		FetchedAt:        s.FetchedAt,
	}
	switch ConditionFromCode(s.Code) {
	case ConditionFog:
		v.FogDensity = 0.8
		v.CloudCoverPct = maxInt(v.CloudCoverPct, 70)
	case ConditionRain:
		if v.PrecipitationMMH == 0 {
			v.PrecipitationMMH = rainFloor(s.Code)
		}
		v.CloudCoverPct = maxInt(v.CloudCoverPct, 80)
	case ConditionSnow:
		if v.SnowfallCMH == 0 {
			v.SnowfallCMH = 0.5
		}
		v.CloudCoverPct = maxInt(v.CloudCoverPct, 80)
	case ConditionStorm:
		v.Thunder = true
		if v.PrecipitationMMH == 0 {
			v.PrecipitationMMH = 4
		}
		v.CloudCoverPct = maxInt(v.CloudCoverPct, 90)
	}
	return v
}

// rainFloor is a rough mm/h for a rain code whose numeric precipitation is 0.
func rainFloor(code int) float64 {
	switch code {
	case 51, 56, 61, 66, 80: // light drizzle, slight rain
		return 0.6
	case 53, 63, 81:
		return 2.5
	case 55, 57, 65, 67, 82:
		return 7
	default:
		return 1
	}
}

func nonNegative(v float64) float64 {
	if v < 0 || v != v {
		return 0
	}
	return v
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
