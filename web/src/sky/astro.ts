// Sun and moon positions from latitude, longitude and time.
//
// The formulas follow the widely used SunCalc algorithm (Vladimir Agafonkin,
// BSD-2-Clause), which is based on "Astronomy Answers" by Louis Strous. They
// are accurate to a fraction of a degree, which is far more than a sky
// animation needs. Everything is computed locally; no network is involved.

const RAD = Math.PI / 180;
const DAY_MS = 86_400_000;
const J1970 = 2_440_588;
const J2000 = 2_451_545;
const OBLIQUITY = RAD * 23.4397; // obliquity of the Earth

const toDays = (date: Date): number => date.valueOf() / DAY_MS - 0.5 + J1970 - J2000;

const rightAscension = (l: number, b: number): number =>
  Math.atan2(Math.sin(l) * Math.cos(OBLIQUITY) - Math.tan(b) * Math.sin(OBLIQUITY), Math.cos(l));

const declination = (l: number, b: number): number =>
  Math.asin(Math.sin(b) * Math.cos(OBLIQUITY) + Math.cos(b) * Math.sin(OBLIQUITY) * Math.sin(l));

// Azimuth in radians measured from the south towards the west (SunCalc convention).
const azimuth = (h: number, phi: number, dec: number): number =>
  Math.atan2(Math.sin(h), Math.cos(h) * Math.sin(phi) - Math.tan(dec) * Math.cos(phi));

const altitude = (h: number, phi: number, dec: number): number =>
  Math.asin(Math.sin(phi) * Math.sin(dec) + Math.cos(phi) * Math.cos(dec) * Math.cos(h));

const siderealTime = (d: number, lw: number): number => RAD * (280.16 + 360.9856235 * d) - lw;

const solarMeanAnomaly = (d: number): number => RAD * (357.5291 + 0.98560028 * d);

function eclipticLongitude(m: number): number {
  const c = RAD * (1.9148 * Math.sin(m) + 0.02 * Math.sin(2 * m) + 0.0003 * Math.sin(3 * m));
  const perihelion = RAD * 102.9372;
  return m + c + perihelion + Math.PI;
}

function sunCoords(d: number): { dec: number; ra: number } {
  const l = eclipticLongitude(solarMeanAnomaly(d));
  return { dec: declination(l, 0), ra: rightAscension(l, 0) };
}

function moonCoords(d: number): { ra: number; dec: number; dist: number } {
  const l = RAD * (218.316 + 13.176396 * d); // ecliptic longitude
  const m = RAD * (134.963 + 13.064993 * d); // mean anomaly
  const f = RAD * (93.272 + 13.22935 * d); // distance from the ascending node
  const lon = l + RAD * 6.289 * Math.sin(m);
  const lat = RAD * 5.128 * Math.sin(f);
  return {
    ra: rightAscension(lon, lat),
    dec: declination(lon, lat),
    dist: 385_001 - 20_905 * Math.cos(m), // km
  };
}

/** Where a body is in the sky. Azimuth is a compass bearing: 0 = north, 90 = east, 180 = south. */
export interface SkyPosition {
  elevationDeg: number;
  azimuthDeg: number;
}

const compass = (azimuthFromSouth: number): number => (azimuthFromSouth / RAD + 180 + 360) % 360;

/** Sun position for a place and moment. Latitude and longitude in degrees (east positive). */
export function sunPosition(date: Date, latDeg: number, lonDeg: number): SkyPosition {
  const lw = RAD * -lonDeg;
  const phi = RAD * latDeg;
  const d = toDays(date);
  const c = sunCoords(d);
  const h = siderealTime(d, lw) - c.ra;
  return {
    elevationDeg: altitude(h, phi, c.dec) / RAD,
    azimuthDeg: compass(azimuth(h, phi, c.dec)),
  };
}

/** Moon position for a place and moment (no atmospheric refraction). */
export function moonPosition(date: Date, latDeg: number, lonDeg: number): SkyPosition {
  const lw = RAD * -lonDeg;
  const phi = RAD * latDeg;
  const d = toDays(date);
  const c = moonCoords(d);
  const h = siderealTime(d, lw) - c.ra;
  return {
    elevationDeg: altitude(h, phi, c.dec) / RAD,
    azimuthDeg: compass(azimuth(h, phi, c.dec)),
  };
}

export interface MoonLight {
  /** Lunar cycle position: 0 = new moon, 0.25 = first quarter, 0.5 = full moon, 0.75 = last quarter. */
  phase: number;
  /** Illuminated fraction of the disc, 0..1. */
  fraction: number;
}

/** Phase and illuminated fraction of the moon. It does not depend on the observer's place. */
export function moonLight(date: Date): MoonLight {
  const d = toDays(date);
  const s = sunCoords(d);
  const m = moonCoords(d);
  const sunDistance = 149_598_000; // km

  const phi = Math.acos(
    Math.sin(s.dec) * Math.sin(m.dec) + Math.cos(s.dec) * Math.cos(m.dec) * Math.cos(s.ra - m.ra),
  );
  const inc = Math.atan2(sunDistance * Math.sin(phi), m.dist - sunDistance * Math.cos(phi));
  const angle = Math.atan2(
    Math.cos(s.dec) * Math.sin(s.ra - m.ra),
    Math.sin(s.dec) * Math.cos(m.dec) - Math.cos(s.dec) * Math.sin(m.dec) * Math.cos(s.ra - m.ra),
  );

  return {
    fraction: (1 + Math.cos(inc)) / 2,
    phase: 0.5 + (0.5 * inc * (angle < 0 ? -1 : 1)) / Math.PI,
  };
}
