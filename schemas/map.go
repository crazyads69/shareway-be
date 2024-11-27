package schemas

import (
	"time"

	"github.com/google/uuid"
)

// Define AutoCompleteRequest struct
type AutoCompleteRequest struct {
	Input           string `form:"input" binding:"required"`
	Location        string `form:"location"`         // Location (latitude,longitude) for which the autocomplete is performed
	Limit           int    `form:"limit"`            // Limit the number of results returned default is 10
	Radius          int    `form:"radius"`           // The distance (in kilometers) within which to return place results (default: 50 km)
	MoreCompound    bool   `form:"more_compound"`    // If true, the API will return more compound results (autocomplete returns fields like district, commune, province. Defaults to false.)
	CurrentLocation string `form:"current_location"` // Current location of the user
}

// Define GoongAutoCompleteResponse struct
type GoongAutoCompleteResponse struct {
	Predictions     []Prediction `json:"predictions"`
	ExecutedTime    int          `json:"executed_time"`
	ExecutedTimeAll int          `json:"executed_time_all"`
	Status          string       `json:"status"`
}

// Define Prediction struct
type Prediction struct {
	Description          string               `json:"description"`
	MatchedSubstrings    []MatchedSubstring   `json:"matched_substrings"`
	PlaceID              string               `json:"place_id"`
	Reference            string               `json:"reference"`
	StructuredFormatting StructuredFormatting `json:"structured_formatting"`
	Terms                []Term               `json:"terms"`
	HasChildren          bool                 `json:"has_children"`
	DisplayType          string               `json:"display_type"`
	Score                float64              `json:"score"`
	PlusCode             PlusCode             `json:"plus_code"`
	Distance             float64              `json:"distance,omitempty"` // Distance from the location (in kilometers) for which the autocomplete is performed
}

// Define MatchedSubstring struct
type MatchedSubstring struct {
	Length int `json:"length"`
	Offset int `json:"offset"`
}

// Define StructuredFormatting struct
type StructuredFormatting struct {
	MainText      string `json:"main_text"`
	SecondaryText string `json:"secondary_text"`
}

// Define Term struct
type Term struct {
	Offset int    `json:"offset"`
	Value  string `json:"value"`
}

// Define PlusCode struct
type PlusCode struct {
	CompoundCode string `json:"compound_code"`
	GlobalCode   string `json:"global_code"`
}

// Define Point struct
type Point struct {
	Lat float64 `json:"lat"` // Latitude
	Lng float64 `json:"lng"` // Longitude
}

// Define GiveRideRequest struct
type GiveRideRequest struct {
	// Points []Point `json:"points" binding:"required"` // List of points for the route
	PlaceList []string  `json:"place_list" binding:"required"`                               // List of places for the route (place_id) from goong api
	StartTime string    `json:"start_time,omitempty"`                                        // Start time of the ride (if not provided, the ride is immediate)
	VehicleID uuid.UUID `json:"vehicle_id" binding:"required,uuid" validate:"required,uuid"` // Vehicle ID for the ride that user has registered
}

// Define

type GoongDirectionsResponse struct {
	Geocoded_waypoints []struct {
		Geocoder_status string `json:"geocoder_status"`
		Place_id        string `json:"place_id"`
	} `json:"geocoded_waypoints"`
	Routes []struct {
		Bounds struct{} `json:"bounds"` // Empty struct as the bounds are empty in the example
		Legs   []struct {
			Distance struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"distance"`
			Duration struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"duration"`
			End_address   string `json:"end_address"`
			Start_address string `json:"start_address"`
			End_location  struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"end_location"`
			Start_location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"start_location"`
			Steps []struct {
				Distance struct {
					Text  string `json:"text"`
					Value int    `json:"value"`
				} `json:"distance"`
				Duration struct {
					Text  string `json:"text"`
					Value int    `json:"value"`
				} `json:"duration"`
				End_location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"end_location"`
				Html_instructions string `json:"html_instructions"`
				Maneuver          string `json:"maneuver"`
				Polyline          struct {
					Points string `json:"points"`
				} `json:"polyline"`
				Start_location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"start_location"`
				Travel_mode string `json:"travel_mode"`
			} `json:"steps"`
		} `json:"legs"`
		Overview_polyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
		Summary        string   `json:"summary"`
		Warnings       []string `json:"warnings"`
		Waypoint_order []int    `json:"waypoint_order"`
	} `json:"routes"`
}

// Define GiveRideResponse struct
type GiveRideResponse struct {
	Route       GoongDirectionsResponse `json:"route"`
	RideOfferID uuid.UUID               `json:"ride_offer_id"`
	Distance    float64                 `json:"distance"`
	Duration    int                     `json:"duration"`
	StartTime   time.Time               `json:"start_time"`
	EndTime     time.Time               `json:"end_time"`
	Fare        float64                 `json:"fare"`
	Vehicle     VehicleDetail           `json:"vehicle"`
	Waypoints   []Waypoint              `json:"waypoints"`
}

// Define HitchRideRequest struct
type HitchRideRequest struct {
	// Points []Point `json:"points" binding:"required"` // List of points for the route
	PlaceList []string `json:"place_list" binding:"required"` // List of places for the route (place_id) from goong api
	StartTime string   `json:"start_time,omitempty"`          // Start time of the ride (if not provided, the ride is immediate)
}

// Define HitchRideResponse struct
type HitchRideResponse struct {
	Route         GoongDirectionsResponse `json:"route"`
	RideRequestID uuid.UUID               `json:"ride_request_id"`
	StartTime     time.Time               `json:"start_time"`
	EndTime       time.Time               `json:"end_time"`
	Distance      float64                 `json:"distance"`
	Duration      int                     `json:"duration"`
}

// Type GeoCodeRequest struct
type GeoCodeRequest struct {
	Point           Point `json:"point" binding:"required"`            // Point for which the geocode is performed
	CurrentLocation Point `json:"current_location" binding:"required"` // Current location of the user
}

// Define GoongGeocodeResponse struct
type GoongReverseGeocodeResponse struct {
	PlusCode PlusCode `json:"plus_code"`
	Results  []Result `json:"results"`
	Status   string   `json:"status"`
}

type Result struct {
	AddressComponents []AddressComponent `json:"address_components"`
	FormattedAddress  string             `json:"formatted_address"`
	Geometry          Geometry           `json:"geometry"`
	PlaceID           string             `json:"place_id"`
	Reference         string             `json:"reference"`
	PlusCode          PlusCode           `json:"plus_code"`
	Types             []string           `json:"types"`
}

type AddressComponent struct {
	LongName  string `json:"long_name"`
	ShortName string `json:"short_name"`
}

type Geometry struct {
	Location Location `json:"location"`
	Boundary string   `json:"boundary,omitempty"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Define GeoCodeLocationResponse struct for the response of the geocode location endpoint
type GeoCodeLocation struct {
	PlaceID          string  `json:"place_id"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	MainAddress      string  `json:"main_address"` // Main address of the location
	SecondaryAddress string  `json:"secondary_address"`
	Distance         float64 `json:"distance,omitempty"` // Distance from the location (in kilometers) for which the geocode is performed
}

// Define GeoCodeLocationResponse struct
type GeoCodeLocationResponse struct {
	Results []GeoCodeLocation `json:"results"`
}

// Define GoongPlaceDetailResponse struct
type GoongPlaceDetailResponse struct {
	Result PlaceDetail `json:"result"`
	Status string      `json:"status"`
}

type PlaceDetail struct {
	PlaceID          string   `json:"place_id"`
	FormattedAddress string   `json:"formatted_address"`
	Geometry         Geometry `json:"geometry"`
	Name             string   `json:"name"`
}

// Define GoongDistanceMatrixResponse struct
type GoongDistanceMatrixResponse struct {
	Rows []struct {
		Elements []struct {
			Status   string `json:"status"`
			Duration struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"duration"`
			Distance struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"distance"`
		} `json:"elements"`
	} `json:"rows"`
}

// Define SuggestRideRequestRequest struct
type SuggestRideRequestRequest struct {
	RideOfferID uuid.UUID `json:"ride_offer_id" binding:"required,uuid" validate:"required,uuid"` // Ride offer ID for which the user wants to suggest a ride request
}

// Define SuggestRideRequestResponse struct
type SuggestRideRequestResponse struct {
	RideRequests []RideRequestDetail `json:"ride_requests"`
}

// Define UserInfo struct
type UserInfo struct {
	ID          uuid.UUID `json:"user_id"`
	PhoneNumber string    `json:"phone_number"`
	FullName    string    `json:"full_name"`
	AvatarURL   string    `json:"avatar_url"`
	Gender      string    `json:"gender"`
}

// Define RideRequestDetail struct
type RideRequestDetail struct {
	ID                    uuid.UUID `json:"ride_request_id"`
	User                  UserInfo  `json:"user"`
	StartLatitude         float64   `json:"start_latitude"`
	StartLongitude        float64   `json:"start_longitude"`
	EndLatitude           float64   `json:"end_latitude"`
	EndLongitude          float64   `json:"end_longitude"`
	RiderCurrentLatitude  float64   `json:"rider_current_latitude"`
	RiderCurrentLongitude float64   `json:"rider_current_longitude"`
	StartAddress          string    `json:"start_address"`
	EndAddress            string    `json:"end_address"`
	Status                string    `json:"status"`
	EncodedPolyline       string    `json:"encoded_polyline"`
	Distance              float64   `json:"distance"`
	Duration              int       `json:"duration"`
	StartTime             time.Time `json:"start_time"`
	EndTime               time.Time `json:"end_time"`
}

// Define SuggestRideOfferRequest struct
type SuggestRideOfferRequest struct {
	RideRequestID uuid.UUID `json:"ride_request_id" binding:"required,uuid" validate:"required,uuid"` // Ride request ID for which the user wants to suggest a ride offer
}

// Define SuggestRideOfferResponse struct
type SuggestRideOfferResponse struct {
	RideOffers []RideOfferDetail `json:"ride_offers"`
}

// Define Waypoint struct
type Waypoint struct {
	ID        uuid.UUID `json:"waypoint_id"`
	Latitude  float64   `json:"lattitude"`
	Longitude float64   `json:"longitude"`
	Address   string    `json:"address"`
	Order     int       `json:"order"`
}

// Define RideOfferDetail struct
type RideOfferDetail struct {
	ID                     uuid.UUID     `json:"ride_offer_id"`
	User                   UserInfo      `json:"user"`
	Vehicle                VehicleDetail `json:"vehicle"`
	StartLatitude          float64       `json:"start_latitude"`
	StartLongitude         float64       `json:"start_longitude"`
	EndLatitude            float64       `json:"end_latitude"`
	EndLongitude           float64       `json:"end_longitude"`
	StartAddress           string        `json:"start_address"`
	EndAddress             string        `json:"end_address"`
	EncodedPolyline        string        `json:"encoded_polyline"`
	Distance               float64       `json:"distance"`
	Duration               int           `json:"duration"`
	DriverCurrentLatitude  float64       `json:"driver_current_latitude"`
	DriverCurrentLongitude float64       `json:"driver_current_longitude"`
	StartTime              time.Time     `json:"start_time"`
	EndTime                time.Time     `json:"end_time"`
	Status                 string        `json:"status"`
	Fare                   float64       `json:"fare"`
	Waypoints              []Waypoint    `json:"waypoints"`
}
