package migration

import (
	"shareway/util/jsonb"
	"shareway/util/polyline"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	// DeletedAt         gorm.DeletedAt `gorm:"index"`
	PhoneNumber       string `gorm:"uniqueIndex;not null"`
	Email             string
	CCCDNumber        string
	FullName          string
	IsVerified        bool `gorm:"default:false"`
	IsActivated       bool `gorm:"default:false"` // Only activated user when first registered and verified OTP completely
	VerifiedAt        time.Time
	Role              string             `gorm:"default:'user'"`
	DeviceToken       string             // FCM token for push notification
	Vehicles          []Vehicle          // One-to-many relationship with Vehicle
	RatingsReceived   []Rating           `gorm:"foreignKey:RateeID"` // One-to-many relationship with Rating (received)
	RatingsGiven      []Rating           `gorm:"foreignKey:RaterID"` // One-to-many relationship with Rating (given)
	RideRequests      []RideRequest      // One-to-many relationship with RideRequest
	RideOffers        []RideOffer        // One-to-many relationship with RideOffer
	Notifications     []Notification     // One-to-many relationship with Notification
	FavoriteLocations []FavoriteLocation // One-to-many relationship with FavoriteLocation
	SentChats         []Chat             `gorm:"foreignKey:SenderID"`   // One-to-many relationship with Chat (sent)
	ReceivedChats     []Chat             `gorm:"foreignKey:ReceiverID"` // One-to-many relationship with Chat (received)
}

// Admin represents an administrator in the system
type Admin struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	// DeletedAt gorm.DeletedAt `gorm:"index"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	FullName string
	Role     string         `gorm:"default:'admin'"`
	Tokens   []SanctumToken `gorm:"foreignKey:AdminID"` // Add reverse relation

}

// SanctumToken represents a Sanctum token for user authentication (for admin)
type SanctumToken struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	AdminID   uuid.UUID `gorm:"type:uuid"`
	Admin     Admin     `gorm:"foreignKey:AdminID"`
	Token     string
	ExpiredAt time.Time
	IsRevoked bool   `gorm:"default:false"` // Revoke the token if needed to prevent further usage
	Ability   string `gorm:"default:'*'"`   // Default is "*" for admin
}

// OTP represents a one-time password for user verification
type OTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	// DeletedAt   gorm.DeletedAt `gorm:"index"`
	PhoneNumber string
	Code        string
	Retry       int `gorm:"default:0"` // Max 3 retries
	ExpiresAt   time.Time
	UserID      uuid.UUID `gorm:"type:uuid"` // Foreign key to User
	User        User      `gorm:"foreignKey:UserID"`
}

// PasetoToken represents a PASETO token for user authentication
type PasetoToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	// DeletedAt    gorm.DeletedAt `gorm:"index"`
	UserID       uuid.UUID `gorm:"type:uuid"`
	User         User      `gorm:"foreignKey:UserID"`
	AccessToken  string    `gorm:"type:text"`
	RefreshToken string    `gorm:"type:text"`
	Revoke       bool      `gorm:"default:false"`
	RefreshTurns int       `gorm:"default:0"` // Max 3 refreshes per access tokee
}

// Transaction represents a payment transaction
type Transaction struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
	PayerID       uuid.UUID `gorm:"type:uuid"`
	Payer         User      `gorm:"foreignKey:PayerID"`
	ReceiverID    uuid.UUID `gorm:"type:uuid"`
	Receiver      User      `gorm:"foreignKey:ReceiverID"`
	Amount        float64
	PaymentMethod string    `gorm:"default:'cash'"`    // cash, wallet, credit_card
	Status        string    `gorm:"default:'pending'"` // pending, completed, failed, refunded
	RideID        uuid.UUID `gorm:"type:uuid"`
	Ride          Ride      `gorm:"foreignKey:RideID"`
}

// Vehicle represents a vehicle in the system
type Vehicle struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt     time.Time   `gorm:"autoCreateTime"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime"`
	LicensePlate  string      `gorm:"uniqueIndex"`
	UserID        uuid.UUID   `gorm:"type:uuid"`
	User          User        `gorm:"foreignKey:UserID"`
	VehicleTypeID uuid.UUID   `gorm:"type:uuid"`
	VehicleType   VehicleType `gorm:"foreignKey:VehicleTypeID"`
	Name          string
	CaVet         string      `gorm:"uniqueIndex"` // Certificate of vehicle registration each vehicle has a unique number
	FuelConsumed  float64     `gorm:"default:0"`   // liters per 100 kilometers
	RideOffers    []RideOffer // One-to-many relationship with RideOffer
}

type RideOffer struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt              time.Time `gorm:"autoCreateTime"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime"`
	UserID                 uuid.UUID `gorm:"type:uuid"`
	User                   User      `gorm:"foreignKey:UserID"`
	VehicleID              uuid.UUID `gorm:"type:uuid"`
	Vehicle                Vehicle   `gorm:"foreignKey:VehicleID"`
	StartLatitude          float64
	StartLongitude         float64
	EndLatitude            float64
	EndLongitude           float64
	EncodedPolyline        polyline.Polyline `gorm:"type:text"` // Store the overview_polyline here
	DriverCurrentLatitude  float64
	DriverCurrentLongitude float64
	StartAddress           string  `gorm:"type:text"`
	EndAddress             string  `gorm:"type:text"`
	Distance               float64 // in kilometers
	Duration               int     // in seconds
	Status                 string  `gorm:"default:'created'"` // created, matched, ongoing, completed, cancelled
	Rides                  []Ride  `gorm:"foreignKey:RideOfferID"`
	StartTime              time.Time
	EndTime                time.Time // Time to end the ride (end time = start time + duration)
	Fare                   float64   // Total price of the ride offer (to show to the hitchhiker)
}

// RideRequest represents a ride request in the system
type RideRequest struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime"`
	UserID                uuid.UUID `gorm:"type:uuid"` // user who requested the ride
	User                  User      `gorm:"foreignKey:UserID"`
	StartLatitude         float64
	StartLongitude        float64
	EndLatitude           float64
	EndLongitude          float64
	RiderCurrentLatitude  float64
	RiderCurrentLongitude float64
	StartAddress          string            `gorm:"type:text"`
	EndAddress            string            `gorm:"type:text"`
	Status                string            `gorm:"default:'created'"` // created, matched, ongoing, completed, cancelled
	Rides                 []Ride            `gorm:"foreignKey:RideRequestID"`
	EncodedPolyline       polyline.Polyline `gorm:"type:text"`
	Distance              float64           // in kilometers
	Duration              int               // in seconds
	StartTime             time.Time
	EndTime               time.Time // Time to end the ride (end time = start time + duration)
}

// Ride represents a matched ride between an offer and a request
type Ride struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt       time.Time   `gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime"`
	RideOfferID     uuid.UUID   `gorm:"type:uuid"`
	RideOffer       RideOffer   `gorm:"foreignKey:RideOfferID"`
	RideRequestID   uuid.UUID   `gorm:"type:uuid"`
	RideRequest     RideRequest `gorm:"foreignKey:RideRequestID"`
	Status          string      `gorm:"default:'scheduled'"` // scheduled, ongoing, completed, cancelled
	StartTime       time.Time
	EndTime         time.Time
	Fare            float64
	StartAddress    string            `gorm:"type:text"`
	EndAddress      string            `gorm:"type:text"`
	EncodedPolyline polyline.Polyline `gorm:"type:text"`
	Distance        float64
	Duration        int
	StartLatitude   float64
	StartLongitude  float64
	EndLatitude     float64
	EndLongitude    float64
	VehicleID       uuid.UUID     `gorm:"type:uuid"`
	Vehicle         Vehicle       `gorm:"foreignKey:VehicleID"`
	Transactions    []Transaction `gorm:"foreignKey:RideID"`
	Ratings         []Rating      `gorm:"foreignKey:RideID"`
}

// Rating represents a rating given by a user to another user
type Rating struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Rating    float64   `gorm:"default:0;check:rating >= 0 AND rating <= 5"`
	Comment   string
	RaterID   uuid.UUID `gorm:"type:uuid"` // user who gave the rating
	Rater     User      `gorm:"foreignKey:RaterID"`
	RateeID   uuid.UUID `gorm:"type:uuid"` // user who received the rating
	Ratee     User      `gorm:"foreignKey:RateeID"`
	RideID    uuid.UUID `gorm:"type:uuid"`
	Ride      Ride      `gorm:"foreignKey:RideID"`
}

// Notification represents a notification sent to a user
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	User      User      `gorm:"foreignKey:UserID"`
	Title     string
	Body      string
	Data      jsonb.JSONB `gorm:"type:jsonb"` // Additional data to be sent with the notification (optional)
	TokenFCM  string      // FCM token of the user to send the notification
	IsRead    bool        `gorm:"default:false"`
}

// Chat represents a chat message between 2 users
type Chat struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	SenderID    uuid.UUID `gorm:"type:uuid"`
	Sender      User      `gorm:"foreignKey:SenderID"`
	ReceiverID  uuid.UUID `gorm:"type:uuid"`
	Receiver    User      `gorm:"foreignKey:ReceiverID"`
	Message     string
	MessageType string    `gorm:"default:'text'"` // text, image, call, missed_call
	RoomID      uuid.UUID `gorm:"type:uuid"`
	Room        Room      `gorm:"foreignKey:RoomID"`
}

// Room represents a chat room between 2 users (1-1 chat)
type Room struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
	User1ID         uuid.UUID `gorm:"type:uuid;index"`
	User1           User      `gorm:"foreignKey:User1ID"`
	User2ID         uuid.UUID `gorm:"type:uuid;index"`
	User2           User      `gorm:"foreignKey:User2ID"`
	LastMessageAt   time.Time `gorm:"index"`     // Add this for sorting/querying recent chats
	LastMessageText string    `gorm:"type:text"` // Cache last message for preview
	LastMessageID   uuid.UUID `gorm:"type:uuid"` // Cache last message ID for preview
	Chats           []Chat    `gorm:"foreignKey:RoomID"`
}

// FavoriteLocation represents a favorite location saved by a user
type FavoriteLocation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	User      User      `gorm:"foreignKey:UserID"`
	Name      string
	Latitude  float64
	Longitude float64
}

// FuelPrice represents the price of fuel
type FuelPrice struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	FuelType  string    `gorm:"uniqueIndex"`
	Price     float64
}

// VehicleType represents a type of vehicle in the system
type VehicleType struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	Name         string    `gorm:"uniqueIndex"`
	FuelConsumed float64   `gorm:"default:0"` // liters per 100 kilometers
	Vehicles     []Vehicle // One-to-many relationship with Vehicle
}
