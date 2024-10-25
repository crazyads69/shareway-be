package schemas

// Define AutoCompleteRequest struct
type AutoCompleteRequest struct {
	Input        string `form:"input" binding:"required"`
	Location     string `form:"location"`      // Location (latitude,longitude) for which the autocomplete is performed
	Limit        int    `form:"limit"`         // Limit the number of results returned default is 10
	Radius       int    `form:"radius"`        // The distance (in kilometers) within which to return place results (default: 50 km)
	MoreCompound bool   `form:"more_compound"` // If true, the API will return more compound results (autocomplete returns fields like district, commune, province. Defaults to false.)
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
	Points []Point `json:"points" binding:"required"` // List of points for the route
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

// Define HitchRideRequest struct
type HitchRideRequest struct {
	Points []Point `json:"points" binding:"required"` // List of points for the route
}
