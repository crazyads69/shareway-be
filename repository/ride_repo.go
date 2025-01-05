package repository

import (
	"errors"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IRideRepository interface {
	CreateNewChatRoom(userID1, userID2 uuid.UUID) error
	GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error)
	InitFirstMessage(roomID uuid.UUID, senderID uuid.UUID, message string) error
	GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error)
	GetTransactionByRideID(rideID uuid.UUID) (migration.Transaction, error)
	AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error)
	CreateRideTransaction(rideID uuid.UUID, Fare float64, paymentMethod string, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error)
	StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error)
	EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error)
	UpdateRideLocation(req schemas.UpdateRideLocationRequest, userID uuid.UUID) (migration.Ride, error)
	CancelRide(req schemas.CancelRideRequest, userID uuid.UUID) (migration.Ride, error)
	GetAllPendingRide(userID uuid.UUID) ([]migration.RideOffer, []migration.RideRequest, error)
}

type RideRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRideRepository(db *gorm.DB, redis *redis.Client) IRideRepository {
	return &RideRepository{db: db, redis: redis}
}

var (
	ErrRideOfferNotFound   = errors.New("ride offer not found")
	ErrRideRequestNotFound = errors.New("ride request not found")
)

// CreateNewChatRoom creates a new chat room between two users
func (r *RideRepository) CreateNewChatRoom(userID1, userID2 uuid.UUID) error {
	// Create a new chat room
	var chatRoom migration.Room
	// Ensure user IDs are in a consistent order
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}

	// Check if the chat room already exists
	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ? AND user2_id = ?", userID1, userID2).
		Or("user1_id = ? AND user2_id = ?", userID2, userID1).
		First(&chatRoom).Error
	if err == nil {
		return nil
	}
	// Create the chat room
	chatRoom = migration.Room{
		User1ID:         userID1,
		User2ID:         userID2,
		LastMessageAt:   time.Now().UTC(), // Set to current time
		LastMessageText: "Hai bạn đã được kết nối. Bắt đầu trò chuyện!",
		LastMessageID:   uuid.New(),
	}
	if err := r.db.Create(&chatRoom).Error; err != nil {
		return err
	}

	return nil
}

// InitFirstMessage initializes the first message in a chat room
func (r *RideRepository) InitFirstMessage(roomID uuid.UUID, senderID uuid.UUID, message string) error {
	// Create a new message
	newMessage := migration.Chat{
		RoomID:      roomID,
		SenderID:    senderID,
		Message:     message,
		MessageType: "text",
	}

	// Create the message
	if err := r.db.Create(&newMessage).Error; err != nil {
		return err
	}

	// Update the chat room with the last message ID and message content
	if err := r.db.Model(&migration.Room{}).Where("id = ?", roomID).Updates(map[string]interface{}{
		"last_message_id":   newMessage.ID,
		"last_message_text": message,
		"last_message_time": newMessage.CreatedAt,
	}).Error; err != nil {
		return err
	}

	return nil
}

// GetChatRoomByUserIDs fetches a chat room by user IDs
func (r *RideRepository) GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error) {
	var chatRoom migration.Room
	// Ensure user IDs are in a consistent order
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}

	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ? AND user2_id = ?", userID1, userID2).
		Or("user1_id = ? AND user2_id = ?", userID2, userID1).
		First(&chatRoom).
		Error

	if err != nil {
		return migration.Room{}, err
	}

	return chatRoom, nil
}

// GetRideOfferByID fetches a ride offer by its ID
func (r *RideRepository) GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error) {
	var rideOffer migration.RideOffer
	err := r.db.Model(&migration.RideOffer{}).
		Select("*"). // Replace with specific fields if you don't need all
		Where("id = ?", rideOfferID).
		Take(&rideOffer).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rideOffer, ErrRideOfferNotFound
		}
		return rideOffer, err
	}

	return rideOffer, nil
}

// GetRideRequestByID fetches a ride request by its ID
func (r *RideRepository) GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error) {
	var rideRequest migration.RideRequest
	err := r.db.Model(&migration.RideRequest{}).
		Select("*"). // Replace with specific fields if you don't need all
		Where("id = ?", rideRequestID).
		Take(&rideRequest).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rideRequest, ErrRideRequestNotFound
		}
		return rideRequest, err
	}

	return rideRequest, nil
}

// AcceptGiveRideRequest accepts a give ride request
func (r *RideRepository) AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride offer by ID with only necessary fields
		var rideOffer migration.RideOffer
		err := tx.Select("user_id, start_time, end_time, status, fare, start_address, end_address, encoded_polyline, distance, duration, start_latitude, start_longitude, end_latitude, end_longitude").
			Where("id = ?", rideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride offer is already matched
		// if rideOffer.Status == "matched" {
		// 	return errors.New("ride offer is already matched")
		// }

		// Get the ride request by ID with only necessary fields
		var rideRequest migration.RideRequest
		err = tx.Select("user_id, start_time, end_time, status, start_address, end_address, start_latitude, start_longitude, end_latitude, end_longitude, encoded_polyline, distance, duration").
			Where("id = ?", rideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride request is already matched
		// if rideRequest.Status == "matched" {
		// 	return errors.New("ride request is already matched")
		// }

		// Create a new ride
		ride = migration.Ride{
			RideOfferID:     rideOfferID,
			RideRequestID:   rideRequestID,
			Status:          "scheduled",
			StartTime:       rideOffer.StartTime,
			EndTime:         rideOffer.EndTime,
			Fare:            rideOffer.Fare,
			StartAddress:    rideOffer.StartAddress,
			EndAddress:      rideOffer.EndAddress,
			EncodedPolyline: rideOffer.EncodedPolyline,
			Distance:        rideOffer.Distance,
			Duration:        rideOffer.Duration,
			StartLatitude:   rideOffer.StartLatitude,
			StartLongitude:  rideOffer.StartLongitude,
			EndLatitude:     rideOffer.EndLatitude,
			EndLongitude:    rideOffer.EndLongitude,
			VehicleID:       vehicleID,
		}

		// Create the ride
		if err := tx.Create(&ride).Error; err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING

		// // Update ride offer status
		// if err := tx.Model(&migration.RideOffer{}).Where("id = ?", rideOfferID).Update("status", "matched").Error; err != nil {
		// 	return err
		// }

		// // Update ride request status
		// if err := tx.Model(&migration.RideRequest{}).Where("id = ?", rideRequestID).Update("status", "matched").Error; err != nil {
		// 	return err
		// }

		// Before creating the chat room, verify both users exist
		var userCount int64
		err = tx.Model(&migration.User{}).
			Where("id IN ?", []uuid.UUID{rideOffer.UserID, rideRequest.UserID}).
			Count(&userCount).Error
		if err != nil {
			return err
		}
		if userCount != 2 {
			return errors.New("one or both users do not exist")
		}

		// Create new chat room (between the driver and the hitcher of the ride)
		// When a ride is accepted, a chat room is created between the driver and the hitcher of the ride
		// then system automatically sends a message to the chat room to notify the hitcher that the ride is accepted
		if err := r.CreateNewChatRoom(rideOffer.UserID, rideRequest.UserID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// CreateRideTransaction creates a transaction for a ride
func (r *RideRepository) CreateRideTransaction(rideID uuid.UUID, Fare float64, paymentMethod string, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error) {
	var transaction migration.Transaction

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create a new transaction
		transaction = migration.Transaction{
			RideID:        rideID,
			Amount:        Fare,
			Status:        "pending",
			PaymentMethod: paymentMethod,
			PayerID:       payerID,
			ReceiverID:    receiverID,
		}

		// Create the transaction
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Transaction{}, err
	}

	return transaction, nil
}

// StartRide starts a ride
func (r *RideRepository) StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride by ID
		err := tx.Model(&migration.Ride{}).
			Where("id = ?", req.RideID).
			First(&ride).Error
		if err != nil {
			return err
		}

		// Get the ride offer by ID
		var rideOffer migration.RideOffer
		err = tx.Model(&migration.RideOffer{}).
			Where("id = ?", ride.RideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}

		// Get the ride request by ID
		var rideRequest migration.RideRequest
		err = tx.Model(&migration.RideRequest{}).
			Where("id = ?", ride.RideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride is already started
		// if ride.Status == "ongoing" {
		// 	return errors.New("ride is already started")
		// }

		// // Check if the ride is already ended
		// if ride.Status == "completed" {
		// 	return errors.New("ride is already ended")
		// }

		// // Check if the ride is already cancelled
		// if ride.Status == "cancelled" {
		// 	return errors.New("ride is already cancelled")
		// }

		// Check if the current location of the driver and hitcher is near less than 100 meters
		if !helper.IsNearby(schemas.Point{Lat: rideOffer.DriverCurrentLatitude, Lng: rideOffer.DriverCurrentLongitude}, schemas.Point{Lat: rideRequest.RiderCurrentLatitude, Lng: rideRequest.RiderCurrentLongitude}, 0.0001) {
			return errors.New("driver and rider are not nearby") // Make sure cannot fake the location
		}

		// TODO: In the future must check start time and end time of the ride to prevent early start or late start

		// Update the ride offer status to ongoing
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", ride.RideOfferID).Update("status", "ongoing").Error; err != nil {
			return err
		}

		// Update the ride request status to ongoing
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", ride.RideRequestID).Update("status", "ongoing").Error; err != nil {
			return err
		}

		// Update the ride status to started
		if err := tx.Model(&migration.Ride{}).Where("id = ?", req.RideID).Update("status", "ongoing").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// GetTransactionByID fetches a transaction by its ID
func (r *RideRepository) GetTransactionByRideID(rideID uuid.UUID) (migration.Transaction, error) {
	var transaction migration.Transaction
	err := r.db.Model(&migration.Transaction{}).
		Where("ride_id = ?", rideID).
		First(&transaction).
		Error

	if err != nil {
		return migration.Transaction{}, err
	}

	return transaction, nil
}

// EndRide ends a ride
func (r *RideRepository) EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride by ID
		err := tx.Model(&migration.Ride{}).
			Where("id = ?", req.RideID).
			First(&ride).Error
		if err != nil {
			return err
		}

		// Get the ride offer by ID
		var rideOffer migration.RideOffer
		err = tx.Model(&migration.RideOffer{}).
			Where("id = ?", ride.RideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}

		// Get the ride request by ID
		var rideRequest migration.RideRequest
		err = tx.Model(&migration.RideRequest{}).
			Where("id = ?", ride.RideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride is already ended
		// if ride.Status == "completed" {
		// 	return errors.New("ride is already ended")
		// }

		// // Check if the ride is already cancelled
		// if ride.Status == "cancelled" {
		// 	return errors.New("ride is already cancelled")
		// }

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// Check if the current location of the driver and the hitcher end location is near less than 100 meters
		// if !helper.IsNearby(schemas.Point{Lat: rideOffer.DriverCurrentLatitude, Lng: rideOffer.DriverCurrentLongitude}, schemas.Point{Lat: rideRequest.EndLatitude, Lng: rideRequest.EndLongitude}, 0.0001) {
		// 	return errors.New("driver not nearby the end location") // Make sure cannot fake the location
		// }

		// Update the ride offer status to ended
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", ride.RideOfferID).Update("status", "completed").Error; err != nil {
			return err
		}

		// Update the ride request status to ended
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", ride.RideRequestID).Update("status", "completed").Error; err != nil {
			return err
		}

		// Update the transaction status to completed
		if err := tx.Model(&migration.Transaction{}).Where("ride_id = ?", req.RideID).Update("status", "completed").Error; err != nil {
			return err
		}

		// Update the ride status to ended
		if err := tx.Model(&migration.Ride{}).Where("id = ?", req.RideID).Update("status", "completed").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// UpdateRideLocation updates the location of a ride
func (r *RideRepository) UpdateRideLocation(req schemas.UpdateRideLocationRequest, userID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride by ID
		err := tx.Model(&migration.Ride{}).
			Where("id = ?", req.RideID).
			First(&ride).Error
		if err != nil {
			return err
		}

		// Get the ride offer by ID
		var rideOffer migration.RideOffer
		err = tx.Model(&migration.RideOffer{}).
			Where("id = ?", ride.RideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}

		// Get the ride request by ID
		var rideRequest migration.RideRequest
		err = tx.Model(&migration.RideRequest{}).
			Where("id = ?", ride.RideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride is already ended
		// if ride.Status == "completed" {
		// 	return errors.New("ride is already ended")
		// }

		// // Check if the ride is already cancelled
		// if ride.Status == "cancelled" {
		// 	return errors.New("ride is already cancelled")
		// }

		// Update the driver's current location
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", ride.RideOfferID).Update("driver_current_latitude", req.CurrentLocation.Lat).Error; err != nil {
			return err
		}
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", ride.RideOfferID).Update("driver_current_longitude", req.CurrentLocation.Lng).Error; err != nil {
			return err
		}

		// Update the hitcher's current location
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", ride.RideRequestID).Update("rider_current_latitude", req.CurrentLocation.Lat).Error; err != nil {
			return err
		}
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", ride.RideRequestID).Update("rider_current_longitude", req.CurrentLocation.Lng).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// CancelRideByDriver cancels a ride by the driver
func (r *RideRepository) CancelRide(req schemas.CancelRideRequest, userID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride by ID
		err := tx.Model(&migration.Ride{}).
			Where("id = ?", req.RideID).
			First(&ride).Error
		if err != nil {
			return err
		}

		// Get the ride offer by ID
		var rideOffer migration.RideOffer
		err = tx.Model(&migration.RideOffer{}).
			Where("id = ?", ride.RideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}

		// Get the ride request by ID
		var rideRequest migration.RideRequest
		err = tx.Model(&migration.RideRequest{}).
			Where("id = ?", ride.RideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}

		// TODO: COMMENTED OUT FOR NOW FOR BETTER TESTING
		// // Check if the ride is already ended
		// if ride.Status == "completed" {
		// 	return errors.New("ride is already ended")
		// }

		// // Check if the ride is already cancelled
		// if ride.Status == "cancelled" {
		// 	return errors.New("ride is already cancelled")
		// }

		// Update the ride offer status to cancelled
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", ride.RideOfferID).Update("status", "cancelled").Error; err != nil {
			return err
		}

		// Update the ride request status to cancelled
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", ride.RideRequestID).Update("status", "cancelled").Error; err != nil {
			return err
		}

		// Update the ride status to cancelled
		if err := tx.Model(&migration.Ride{}).Where("id = ?", req.RideID).Update("status", "cancelled").Error; err != nil {
			return err
		}

		// Update the transaction status to cancelled
		if err := tx.Model(&migration.Transaction{}).Where("ride_id = ?", req.RideID).Update("status", "cancelled").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// GetAllPendingRide fetches all pending rides for a user
func (r *RideRepository) GetAllPendingRide(userID uuid.UUID) ([]migration.RideOffer, []migration.RideRequest, error) {
	var rideOffers []migration.RideOffer
	var rideRequests []migration.RideRequest

	// Get all ride offers that not have status cancelled or completed and not matched (status = created)
	err := r.db.Model(&migration.RideOffer{}).
		Where("user_id = ? AND status = 'created'", userID).
		Find(&rideOffers).
		Error
	if err != nil {
		return nil, nil, err
	}

	// Get all ride requests that not have status cancelled or completed and not matched (status = created)
	err = r.db.Model(&migration.RideRequest{}).
		Where("user_id = ? AND status = 'created'", userID).
		Find(&rideRequests).
		Error
	if err != nil {
		return nil, nil, err
	}

	return rideOffers, rideRequests, nil
}

// Make sure the RideRepository implements the IRideRepository interface
var _ IRideRepository = (*RideRepository)(nil)
