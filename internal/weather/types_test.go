package weather

import "testing"

func TestConditionFromCode(t *testing.T) {
	cases := map[int]Condition{
		0: ConditionClear, 1: ConditionClear, 2: ConditionCloudy, 3: ConditionCloudy,
		45: ConditionFog, 48: ConditionFog, 51: ConditionRain, 65: ConditionRain, 67: ConditionRain, 82: ConditionRain,
		71: ConditionSnow, 77: ConditionSnow, 86: ConditionSnow, 95: ConditionStorm, 99: ConditionStorm, 42: ConditionUnknown,
	}
	for code, want := range cases {
		if got := ConditionFromCode(code); got != want {
			t.Errorf("code %d = %q, want %q", code, got, want)
		}
	}
}

func TestViewFillsGapsFromTheCode(t *testing.T) {
	rain := Snapshot{Code: 63}.View()
	if rain.PrecipitationMMH <= 0 || rain.CloudCoverPct < 80 {
		t.Fatalf("rain code without mm: %+v", rain)
	}
	snow := Snapshot{Code: 73}.View()
	if snow.SnowfallCMH <= 0 {
		t.Fatalf("snow code without cm: %+v", snow)
	}
	if v := (Snapshot{Code: 45}).View(); v.FogDensity < 0.5 {
		t.Fatalf("fog: %+v", v)
	}
	if v := (Snapshot{Code: 95}).View(); !v.Thunder || v.PrecipitationMMH <= 0 {
		t.Fatalf("storm: %+v", v)
	}
	// Real numbers win over the code floor.
	if v := (Snapshot{Code: 63, PrecipitationMMH: 9}).View(); v.PrecipitationMMH != 9 {
		t.Fatalf("measured rain overridden: %+v", v)
	}
	// Bad input is clamped.
	v := Snapshot{CloudCoverPct: 150, WindDirDeg: -90, WindSpeedKMH: -3}.View()
	if v.CloudCoverPct != 100 || v.WindDirDeg != 270 || v.WindSpeedKMH != 0 {
		t.Fatalf("clamping: %+v", v)
	}
}
