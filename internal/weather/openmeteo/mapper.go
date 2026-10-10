package openmeteo

// mapper.go will convert Open-Meteo JSON responses into weather.Snapshot.
// Keeping the mapping separate from the HTTP client makes it testable with
// recorded response fixtures in testdata/.
