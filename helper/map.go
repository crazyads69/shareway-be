package helper

import (
	"math"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"strconv"
	"strings"
	"time"

	"github.com/twpayne/go-polyline"
)

const (
	earthRadius         = 6371   // Earth radius in kilometers
	minDistance         = 0.1    // Minimum distance in kilometers
	maxDistanceSubRoute = 0.0003 // Maximum distance in kilometers
	maxDistanceMatch    = 0.02   // About 2 km at equator
	degreesToRad        = math.Pi / 180
	maxDistanceSq       = maxDistanceMatch * maxDistanceMatch
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

func IsMatchRoute(offerPolyline, requestPolyline []schemas.Point) bool {
	if len(requestPolyline) < 2 || len(offerPolyline) < 2 {
		return false
	}

	if len(requestPolyline) > len(offerPolyline) {
		return false
	}

	if IsSubRoute(offerPolyline, requestPolyline) {
		return true
	}

	startPoint, endPoint := requestPolyline[0], requestPolyline[len(requestPolyline)-1]
	minStartDistSq, minEndDistSq := math.MaxFloat64, math.MaxFloat64
	startIdx, endIdx := -1, -1

	for i, point := range offerPolyline {
		startDistSq := squaredDistance(point, startPoint)
		if startDistSq < minStartDistSq {
			minStartDistSq = startDistSq
			startIdx = i
		}

		endDistSq := squaredDistance(point, endPoint)
		if endDistSq < minEndDistSq {
			minEndDistSq = endDistSq
			endIdx = i
		}
	}

	return minStartDistSq <= maxDistanceSq && minEndDistSq <= maxDistanceSq && startIdx < endIdx
}

func IsSubRoute(offerPolyline, requestPolyline []schemas.Point) bool {
	startPoint, endPoint := requestPolyline[0], requestPolyline[len(requestPolyline)-1]
	startIdx, endIdx := -1, -1

	for i, point := range offerPolyline {
		if startIdx == -1 && squaredDistance(point, startPoint) <= maxDistanceSq {
			startIdx = i
		}
		if squaredDistance(point, endPoint) <= maxDistanceSq {
			endIdx = i
		}
		if startIdx != -1 && endIdx != -1 {
			break
		}
	}

	return startIdx != -1 && endIdx != -1 && startIdx < endIdx
}

func squaredDistance(p1, p2 schemas.Point) float64 {
	dx := (p2.Lng - p1.Lng) * math.Cos((p1.Lat+p2.Lat)/2*degreesToRad)
	dy := p2.Lat - p1.Lat
	return dx*dx + dy*dy
}

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * degreesToRad
	dLon := (lon2 - lon1) * degreesToRad
	lat1 = lat1 * degreesToRad
	lat2 = lat2 * degreesToRad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func IsTimeOverlap(offer migration.RideOffer, request migration.RideRequest) bool {
	// Add a buffer of 30 minutes to the start and end time of the offer
	// to account for the time it takes to pick up the hitchhiker and drop them off
	// This buffer is added to the start and end time of the offer
	offerStartTime := offer.StartTime.Add(-30 * time.Minute)
	offerEndTime := offer.EndTime.Add(30 * time.Minute)
	return offerStartTime.Before(request.StartTime) && offerEndTime.After(request.EndTime)
}

// func IsSubRoute(offerPolyline, requestPolyline []schemas.Point) bool {
// 	if len(requestPolyline) > len(offerPolyline) {
// 		return false
// 	}

// 	for i := 0; i <= len(offerPolyline)-len(requestPolyline); i++ {
// 		if IsMatchingSegment(offerPolyline[i:i+len(requestPolyline)], requestPolyline) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func IsMatchingSegment(segment, requestPolyline []schemas.Point) bool {
// 	const epsilon = 0.0001 // Small threshold for floating-point comparison
// 	for i := range requestPolyline {
// 		if math.Abs(segment[i].Lat-requestPolyline[i].Lat) > epsilon ||
// 			math.Abs(segment[i].Lng-requestPolyline[i].Lng) > epsilon {
// 			return false
// 		}
// 	}
// 	return true
// }

// func IsRouteMatching(offerPolyline, requestPolyline []schemas.Point, maxDistance float64) bool {
// 	// Check if the request route is a subset of the offer route
// 	if IsSubRoute(offerPolyline, requestPolyline) {
// 		return true
// 	}

// 	// Calculate the radius
// 	startDist := HaversineDistance(
// 		offerPolyline[0].Lat, offerPolyline[0].Lng,
// 		requestPolyline[0].Lat, requestPolyline[0].Lng,
// 	)

// 	minEndDist := math.Inf(1)
// 	for i := 0; i < len(offerPolyline); i++ {
// 		endDist := HaversineDistance(
// 			offerPolyline[i].Lat, offerPolyline[i].Lng,
// 			requestPolyline[len(requestPolyline)-1].Lat,
// 			requestPolyline[len(requestPolyline)-1].Lng,
// 		)
// 		if endDist < minEndDist {
// 			minEndDist = endDist
// 		}
// 	}

// 	totalRadius := math.Abs(startDist - minEndDist)
// 	return totalRadius <= maxDistance
// }
