package schemas

import (
	"time"

	"github.com/google/uuid"
)

// There is some duplicate schema but it is for easier for maintain and clean code

// Define SendGiveRideRequestRequest schema
type SendGiveRideRequestRequest struct {
	// The ID of the ride offer (current user is the driver)
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the ride request (the user received request is the hitcher)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the receiver (the user who received the request) aka the hitcher
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	VehicleID  uuid.UUID `json:"vehicleID" binding:"required,uuid" validate:"required,uuid"`
}

// Define SendGiveRideRequestResponse schema
type SendGiveRideRequestResponse struct {
	// This will act as data send through websocket to the receiver to able preview the request before accepting or rejecting
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
	ReceiverID             uuid.UUID     `json:"receiver_id"`
	RideRequestID          uuid.UUID     `json:"ride_request_id"`
	Waypoints              []Waypoint    `json:"waypoints"`
}

// Define SendHitchRideRequestRequest schema
type SendHitchRideRequestRequest struct {
	// The ID of the ride request (current user is the hitcher)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the ride offer (the user who received request is the driver)
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the receiver (the user who received the request) aka the driver
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	VehicleID  uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// Define SendHitchRideRequestResponse schema
type SendHitchRideRequestResponse struct {
	// This will act as data send through websocket to the receiver to able preview the request before accepting or rejecting
	ID                    uuid.UUID     `json:"ride_request_id"`
	User                  UserInfo      `json:"user"`
	Vehicle               VehicleDetail `json:"vehicle"`
	StartLatitude         float64       `json:"start_latitude"`
	StartLongitude        float64       `json:"start_longitude"`
	EndLatitude           float64       `json:"end_latitude"`
	EndLongitude          float64       `json:"end_longitude"`
	RiderCurrentLatitude  float64       `json:"rider_current_latitude"`
	RiderCurrentLongitude float64       `json:"rider_current_longitude"`
	StartAddress          string        `json:"start_address"`
	EndAddress            string        `json:"end_address"`
	Status                string        `json:"status"`
	EncodedPolyline       string        `json:"encoded_polyline"`
	Distance              float64       `json:"distance"`
	Duration              int           `json:"duration"`
	StartTime             time.Time     `json:"start_time"`
	EndTime               time.Time     `json:"end_time"`
	ReceiverID            uuid.UUID     `json:"receiver_id"`
	RideOfferID           uuid.UUID     `json:"ride_offer_id"`
}

// Define AcceptRideGiveRequestRequest schema
type AcceptGiveRideRequestRequest struct {
	// The ID of the ride offer (current user is the driver)
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the ride request (the user received request is the hitcher)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the receiver (the user who received the request) aka the hitcher
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the vehicle
	VehicleID uuid.UUID `json:"vehicleID" binding:"required,uuid" validate:"required,uuid"`
	// Payment method (cash or momo)
	PaymentMethod string `json:"paymentMethod" binding:"required" validate:"required,oneof=cash momo"`
}

type TransactionDetail struct {
	ID            uuid.UUID `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
}

// Define AcceptRideGiveRequestResponse schema
type AcceptGiveRideRequestResponse struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   float64           `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	UserInfo               UserInfo          `json:"user"`
	ReceiverID             uuid.UUID         `json:"receiver_id"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

// Define AcceptHitchRideRequestRequest schema
type AcceptHitchRideRequestRequest struct {
	// The ID of the ride request (current user is the hitcher)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the ride offer (the user who received request is the driver)
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the receiver (the user who received the request) aka the driver
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the vehicle
	VehicleID uuid.UUID `json:"vehicleID" binding:"required,uuid" validate:"required,uuid"`
	// Payment method (cash or momo)
	PaymentMethod string `json:"paymentMethod" binding:"required" validate:"required,oneof=cash momo"`
}

// Define AcceptHitchRideRequestResponse schema
type AcceptHitchRideRequestResponse struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	ReceiverID             uuid.UUID         `json:"receiver_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   float64           `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	UserInfo               UserInfo          `json:"user"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

type CancelGiveRideRequestRequest struct {
	RideOfferID   uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	ReceiverID    uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	VehicleID     uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

type CancelGiveRideRequestResponse struct {
	// Send back id of ride offer ride request to update ui
	RideOfferID   uuid.UUID `json:"ride_offer_id"`
	RideRequestID uuid.UUID `json:"ride_request_id"`
	UserID        uuid.UUID `json:"user_id"` // The user who cancel the request (hitcher cancel the request)
	ReceiverID    uuid.UUID `json:"receiver_id"`
}

type CancelHitchRideRequestRequest struct {

	// The driver who want to cancel the ride request
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The hitcher who received the request
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The hitcher who received the cancel request
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	// The vehicle id
	VehicleID uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

type CancelHitchRideRequestResponse struct {
	// Send back id of ride offer ride request to update ui
	RideOfferID   uuid.UUID `json:"ride_offer_id"`
	RideRequestID uuid.UUID `json:"ride_request_id"`
	UserID        uuid.UUID `json:"user_id"` // The user who cancel the request (driver cancel the request)
	ReceiverID    uuid.UUID `json:"receiver_id"`
}

// Define StartRideRequest schema
type StartRideRequest struct {
	// Ride ID of the ride to start
	RideID uuid.UUID `json:"rideID" binding:"required,uuid" validate:"required,uuid"`
	// Current user location
	CurrentLocation Point     `json:"currentLocation" binding:"required" validate:"required"`
	VehicleID       uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// Define StartRideResponse schema
type StartRideResponse struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	User                   UserInfo          `json:"user"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	ReceiverID             uuid.UUID         `json:"receiver_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   float64           `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

// Define EndRideRequest schema
type EndRideRequest struct {
	// Ride ID of the ride to end
	RideID uuid.UUID `json:"rideID" binding:"required,uuid" validate:"required,uuid"`
	// Current user location
	CurrentLocation Point     `json:"currentLocation" binding:"required" validate:"required"`
	VehicleID       uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// Define EndRideResponse schema
type EndRideResponse struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	User                   UserInfo          `json:"user"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	ReceiverID             uuid.UUID         `json:"receiver_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   float64           `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

// Define UpdateRideLocationRequest schema
type UpdateRideLocationRequest struct {
	// Ride ID of the ride to update location
	RideID uuid.UUID `json:"rideID" binding:"required,uuid" validate:"required,uuid"`
	// Current driver location
	CurrentLocation Point     `json:"currentLocation" binding:"required" validate:"required"`
	VehicleID       uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// Define UpdateRideLocationResponse schema
type UpdateRideLocationResponse struct {
	ID                     uuid.UUID         `json:"ride_id"`
	RideOfferID            uuid.UUID         `json:"ride_offer_id"`
	User                   UserInfo          `json:"user"`
	RideRequestID          uuid.UUID         `json:"ride_request_id"`
	ReceiverID             uuid.UUID         `json:"receiver_id"`
	Status                 string            `json:"status"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                time.Time         `json:"end_time"`
	StartAddress           string            `json:"start_address"`
	EndAddress             string            `json:"end_address"`
	Fare                   float64           `json:"fare"`
	EncodedPolyline        string            `json:"encoded_polyline"`
	Distance               float64           `json:"distance"`
	Duration               int               `json:"duration"`
	Transaction            TransactionDetail `json:"transaction"`
	StartLatitude          float64           `json:"start_latitude"`
	StartLongitude         float64           `json:"start_longitude"`
	EndLatitude            float64           `json:"end_latitude"`
	EndLongitude           float64           `json:"end_longitude"`
	Vehicle                VehicleDetail     `json:"vehicle"`
	DriverCurrentLatitude  float64           `json:"driver_current_latitude"`
	DriverCurrentLongitude float64           `json:"driver_current_longitude"`
	RiderCurrentLatitude   float64           `json:"rider_current_latitude"`
	RiderCurrentLongitude  float64           `json:"rider_current_longitude"`
	Waypoints              []Waypoint        `json:"waypoints"`
}

type CancelRideRequest struct {
	// The ride id to cancel
	RideID uuid.UUID `json:"rideID" binding:"required,uuid" validate:"required,uuid"`
	// The receiver id (the hitcher) who received the cancel request
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	VehicleID  uuid.UUID `json:"vehicleID,omitempty" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

type CancelRideResponse struct {
	// Send back id of ride offer ride request to update ui
	RideID        uuid.UUID `json:"ride_id"`
	ReceiverID    uuid.UUID `json:"receiver_id"`
	RideOfferID   uuid.UUID `json:"ride_offer_id"`
	RideRequestID uuid.UUID `json:"ride_request_id"`
}

// Define GetPendingRide schema
// This schema is used to get the pending ride request and offer of the user
// type GetAllPendingRideRequest struct {
// 	// The ID of the user
// 	UserID uuid.UUID `json:"userID" binding:"required,uuid" validate:"required,uuid"`
// }

// Define GetPendingRideResponse schema
type GetAllPendingRideResponse struct {
	// The pending ride request of the user
	PendingRideRequest []RideRequestDetail `json:"pending_ride_request"`
	// The pending ride offer of the user
	PendingRideOffer []RideOfferDetail `json:"pending_ride_offer"`
}
