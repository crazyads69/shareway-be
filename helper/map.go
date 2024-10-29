package helper

import (
	"math"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"strconv"
	"strings"

	"github.com/twpayne/go-polyline"
)

const (
	earthRadius = 6371 // Earth radius in kilometers
	minDistance = 0.1  // Minimum distance in kilometers
)

func haversineDistance(p1, p2 schemas.Point) float64 {
	dLat := toRadians(p2.Lat - p1.Lat)
	dLon := toRadians(p2.Lng - p1.Lng)
	lat1 := toRadians(p1.Lat)
	lat2 := toRadians(p2.Lat)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func OptimizeRoutePoints(points []schemas.Point) []schemas.Point {
	if len(points) < 2 {
		return points
	}

	optimized := []schemas.Point{points[0]}

	for i := 1; i < len(points); i++ {
		if haversineDistance(optimized[len(optimized)-1], points[i]) >= minDistance {
			optimized = append(optimized, points[i])
		}
	}

	// Always include the last point
	if optimized[len(optimized)-1] != points[len(points)-1] {
		optimized = append(optimized, points[len(points)-1])
	}

	return optimized
}

func DecodePolyline(encodedPolyline string) []schemas.Point {
	buf := []byte(encodedPolyline)
	coords, _, _ := polyline.DecodeCoords(buf)

	var points []schemas.Point
	// the coords is a slice of slices, each slice contains two elements: latitude and longitude
	// and have order to create a route on the map the previous point use for encode the next point
	for _, coord := range coords {
		points = append(points, schemas.Point{Lat: coord[0], Lng: coord[1]})
	}

	return points
}

func ConvertStringToLocation(point string) schemas.Point {
	// Split the string into latitude and longitude
	latLng := strings.Split(point, ",")
	lat, _ := strconv.ParseFloat(latLng[0], 64)
	lng, _ := strconv.ParseFloat(latLng[1], 64)

	return schemas.Point{
		Lat: lat,
		Lng: lng,
	}
}

func IsRouteMatching(offerPolyline, requestPolyline []schemas.Point, maxDistance float64) bool {
	// Check if the request route is a subset of the offer route
	if IsSubRoute(offerPolyline, requestPolyline) {
		return true
	}

	// Calculate the radius
	startDist := HaversineDistance(
		offerPolyline[0].Lat, offerPolyline[0].Lng,
		requestPolyline[0].Lat, requestPolyline[0].Lng,
	)

	minEndDist := math.Inf(1)
	for i := 0; i < len(offerPolyline); i++ {
		endDist := HaversineDistance(
			offerPolyline[i].Lat, offerPolyline[i].Lng,
			requestPolyline[len(requestPolyline)-1].Lat,
			requestPolyline[len(requestPolyline)-1].Lng,
		)
		if endDist < minEndDist {
			minEndDist = endDist
		}
	}

	totalRadius := math.Abs(startDist - minEndDist)
	return totalRadius <= maxDistance
}

func IsSubRoute(offerPolyline, requestPolyline []schemas.Point) bool {
	if len(requestPolyline) > len(offerPolyline) {
		return false
	}

	for i := 0; i <= len(offerPolyline)-len(requestPolyline); i++ {
		if IsMatchingSegment(offerPolyline[i:i+len(requestPolyline)], requestPolyline) {
			return true
		}
	}
	return false
}

func IsMatchingSegment(segment, requestPolyline []schemas.Point) bool {
	const epsilon = 0.0001 // Small threshold for floating-point comparison
	for i := range requestPolyline {
		if math.Abs(segment[i].Lat-requestPolyline[i].Lat) > epsilon ||
			math.Abs(segment[i].Lng-requestPolyline[i].Lng) > epsilon {
			return false
		}
	}
	return true
}

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func IsTimeOverlap(offer migration.RideOffer, request migration.RideRequest) bool {
	return offer.StartTime.Before(request.EndTime) && offer.EndTime.After(request.StartTime)
}
