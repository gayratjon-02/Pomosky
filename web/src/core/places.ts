// Places the host can pick without searching. The server rounds the coordinates.
export interface Place {
  name: string;
  lat: number;
  lon: number;
  tz: string;
}

export const PLACES: Place[] = [
  { name: "Seoul", lat: 37.5665, lon: 126.978, tz: "Asia/Seoul" },
  { name: "Tokyo", lat: 35.6762, lon: 139.6503, tz: "Asia/Tokyo" },
  { name: "Tashkent", lat: 41.2995, lon: 69.2401, tz: "Asia/Tashkent" },
  { name: "Istanbul", lat: 41.0082, lon: 28.9784, tz: "Europe/Istanbul" },
  { name: "London", lat: 51.5072, lon: -0.1276, tz: "Europe/London" },
  { name: "Reykjavik", lat: 64.1466, lon: -21.9426, tz: "Atlantic/Reykjavik" },
  { name: "New York", lat: 40.7128, lon: -74.006, tz: "America/New_York" },
  { name: "Los Angeles", lat: 34.0522, lon: -118.2437, tz: "America/Los_Angeles" },
  { name: "Sao Paulo", lat: -23.5505, lon: -46.6333, tz: "America/Sao_Paulo" },
  { name: "Cape Town", lat: -33.9249, lon: 18.4241, tz: "Africa/Johannesburg" },
  { name: "Singapore", lat: 1.3521, lon: 103.8198, tz: "Asia/Singapore" },
  { name: "Sydney", lat: -33.8688, lon: 151.2093, tz: "Australia/Sydney" },
];

export const DEFAULT_PLACE = PLACES[0]!;
