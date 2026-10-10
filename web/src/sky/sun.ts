// Sun position from latitude, longitude and time.
// TODO(stage 6): implement a NOAA / SunCalc-style algorithm and unit-test it
// against known sunrise and sunset times.

export interface SunPosition {
  elevationDeg: number;
  azimuthDeg: number;
}

export function sunPosition(_date: Date, _latDeg: number, _lonDeg: number): SunPosition {
  return { elevationDeg: 0, azimuthDeg: 0 };
}
