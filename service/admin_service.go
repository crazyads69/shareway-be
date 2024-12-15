package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"shareway/infra/bucket"
	"shareway/infra/db/migration"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"shareway/util/sanctum"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type AdminService struct {
	repo         repository.IAdminRepository
	hub          *ws.Hub
	cfg          util.Config
	cloudinary   *bucket.CloudinaryService
	sanctumToken *sanctum.SanctumToken
}

func NewAdminService(repo repository.IAdminRepository, hub *ws.Hub, cfg util.Config, cloudinary *bucket.CloudinaryService, sanctumToken *sanctum.SanctumToken) IAdminService {
	return &AdminService{
		repo:         repo,
		hub:          hub,
		cfg:          cfg,
		cloudinary:   cloudinary,
		sanctumToken: sanctumToken,
	}
}

type IAdminService interface {
	CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error)
	VerifyPassword(password, hashedPassword string) bool
	CreateToken(admin migration.Admin) (string, error)
	GetAdminProfile(adminID uuid.UUID) (migration.Admin, error)
	GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error)
	GetUserDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.UserDashboardDataResponse, error)
	GetRideDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.RideDashboardDataResponse, error)
	GetTransactionDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.TransactionDashboardDataResponse, error)
	GetVehicleDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.VehicleDashboardDataResponse, error)
	GetUserList(req schemas.UserListRequest) ([]migration.User, int64, int64, error)
	GetRideList(req schemas.RideListRequest) ([]migration.Ride, int64, int64, error)
	GetVehicleList(req schemas.VehicleListRequest) ([]migration.Vehicle, int64, int64, error)
	GetTransactionList(req schemas.TransactionListRequest) ([]migration.Transaction, int64, int64, error)
	GetDashboardData(req schemas.DashboardReportRequest) (schemas.ReportData, error)
	AnalyzeDashboardData(data schemas.ReportData) (string, error)
	CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error)
}

// CheckAdminExists checks if an admin exists with the given email and password
func (s *AdminService) CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error) {
	return s.repo.CheckAdminExists(req)
}

// VerifyPassword verifies if the given password matches the hashed password
func (s *AdminService) VerifyPassword(password, hashedPassword string) bool {
	return s.sanctumToken.Cryto.VerifyPassword(hashedPassword, password)
}

// CreateToken creates a new token for the admin
func (s *AdminService) CreateToken(admin migration.Admin) (string, error) {
	return s.sanctumToken.CreateSanctumToken(admin.ID, time.Duration(s.cfg.RefreshTokenExpiredDuration)*time.Second)
}

// GetAdminProfile gets the profile of the admin
func (s *AdminService) GetAdminProfile(adminID uuid.UUID) (migration.Admin, error) {
	return s.repo.GetAdminProfile(adminID)
}

// GetDashboardGeneralData gets the general data for the dashboard
func (s *AdminService) GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error) {
	return s.repo.GetDashboardGeneralData()
}

// GetUserDashboardData gets the data for the user dashboard
func (s *AdminService) GetUserDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.UserDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.UserDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetUserDashboardData(startDate, endDate)
}

// GetRideDashboardData gets the data for the ride dashboard
func (s *AdminService) GetRideDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.RideDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.RideDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetRideDashboardData(startDate, endDate)
}

// GetTransactionDashboardData gets the data for the transaction dashboard
func (s *AdminService) GetTransactionDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.TransactionDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.TransactionDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetTransactionDashboardData(startDate, endDate)
}

// GetVehicleDashboardData gets the data for the vehicle dashboard
func (s *AdminService) GetVehicleDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.VehicleDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.VehicleDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetVehicleDashboardData(startDate, endDate)
}

// GetUserList gets the list of users
func (s *AdminService) GetUserList(req schemas.UserListRequest) ([]migration.User, int64, int64, error) {
	return s.repo.GetUserList(req)
}

// GetRideList gets the list of rides
func (s *AdminService) GetRideList(req schemas.RideListRequest) ([]migration.Ride, int64, int64, error) {
	return s.repo.GetRideList(req)
}

// GetVehicleList gets the list of vehicles
func (s *AdminService) GetVehicleList(req schemas.VehicleListRequest) ([]migration.Vehicle, int64, int64, error) {
	return s.repo.GetVehicleList(req)
}

// GetTransactionList gets the list of transactions
func (s *AdminService) GetTransactionList(req schemas.TransactionListRequest) ([]migration.Transaction, int64, int64, error) {
	return s.repo.GetTransactionList(req)
}

// GetDashboardData gets the data for the dashboard report
func (s *AdminService) GetDashboardData(req schemas.DashboardReportRequest) (schemas.ReportData, error) {
	return s.repo.GetDashboardData(req)
}

// AnalyzeDashboardData analyzes the data for the dashboard report using OpenRouter AI with LLM
func (s *AdminService) AnalyzeDashboardData(data schemas.ReportData) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following dashboard data and provide insights:
    Total Users: %d
    Active Users: %d
    Total Rides: %d
    Completed Rides: %d
    Cancelled Rides: %d
    Total Transactions Amount: %d VND
    Average Rating: %.2f

    Popular Routes:
    %v

    User Growth:
    %v

    Revenue by Day:
    %v

    Vehicle Type Distribution:
    %v

    Please provide a detailed analysis of the data, including trends, potential areas for improvement, and recommendations for business growth.`,
		data.TotalUsers, data.ActiveUsers, data.TotalRides, data.CompletedRides, data.CancelledRides,
		data.TotalTransactions, data.AverageRating, data.PopularRoutes, data.UserGrowth, data.TransactionByDay,
		data.VehicleTypeDistribution)

	// Call OpenRouter API directly
	requestBody := schemas.OpenRouterRequestBody{
		Model: "google/gemini-exp-1206:free",
		Messages: []schemas.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal the request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create request
	req, err := http.NewRequest("POST", s.cfg.OpenRouterAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.cfg.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response
	var openRouterResponse schemas.OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResponse)
	if err != nil {
		return "", err
	}

	// Extract the response
	if len(openRouterResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	// Extract the response
	response := openRouterResponse.Choices[0].Message.Content

	// Check if the response is a string
	if str, ok := response.(string); ok {
		return str, nil
	} else {
		return "", fmt.Errorf("invalid response from OpenRouter")
	}
}

// CreateExcelReport creates an Excel report from the data and analysis
func (s *AdminService) CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	// Tạo sheet tổng quan
	f.NewSheet("Tổng quan")
	f.SetCellValue("Tổng quan", "A1", "Chỉ số")
	f.SetCellValue("Tổng quan", "B1", "Giá trị")
	f.SetCellValue("Tổng quan", "A2", "Tổng số người dùng")
	f.SetCellValue("Tổng quan", "B2", data.TotalUsers)
	f.SetCellValue("Tổng quan", "A3", "Người dùng hoạt động")
	f.SetCellValue("Tổng quan", "B3", data.ActiveUsers)
	f.SetCellValue("Tổng quan", "A4", "Tông số chuyến đi")
	f.SetCellValue("Tổng quan", "B4", data.TotalRides)
	f.SetCellValue("Tổng quan", "A5", "Số chuyến đi hoàn thành")
	f.SetCellValue("Tổng quan", "B5", data.CompletedRides)
	f.SetCellValue("Tổng quan", "A6", "Số chuyến đi bị hủy")
	f.SetCellValue("Tổng quan", "B6", data.CancelledRides)
	f.SetCellValue("Tổng quan", "A7", "Tổng giá trị giao dịch")
	f.SetCellValue("Tổng quan", "B7", data.TotalTransactions)
	f.SetCellValue("Tổng quan", "A8", "Trung bình đánh giá")
	f.SetCellValue("Tổng quan", "B8", data.AverageRating)

	// Tạo sheet cho các tuyến đường phổ biến
	f.NewSheet("Tuyến đường phổ biến")
	f.SetCellValue("", "A1", "Địa chỉ bắt đầu")
	f.SetCellValue("Tuyến đường phổ biến", "B1", "Địa chỉ kết thúc")
	f.SetCellValue("Tuyến đường phổ biến", "C1", "Tổng số lần")
	for i, route := range data.PopularRoutes {
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("A%d", i+2), route.StartAddress)
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("B%d", i+2), route.EndAddress)
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("C%d", i+2), route.Count)
	}

	// Tạo sheet cho tăng trưởng người dùng
	f.NewSheet("Chỉ số tăng trưởng người dùng")
	f.SetCellValue("Chỉ số tăng trưởng người dùng", "A1", "Ngày")
	f.SetCellValue("Chỉ số tăng trưởng người dùng", "B1", "Số lượng người dùng mới")
	for i, growth := range data.UserGrowth {
		f.SetCellValue("Chỉ số tăng trưởng người dùng", fmt.Sprintf("A%d", i+2), growth.Date.Format("02/01/2006 15:04"))
		f.SetCellValue("Chỉ số tăng trưởng người dùng", fmt.Sprintf("B%d", i+2), growth.Count)
	}

	// Tạo sheet cho giao dịch theo ngày
	f.NewSheet("Giao dịch theo ngày")
	f.SetCellValue("Giao dịch theo ngày", "A1", "Ngày")
	f.SetCellValue("Giao dịch theo ngày", "B1", "Tổng giá trị")
	for i, revenue := range data.TransactionByDay {
		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("A%d", i+2), revenue.Date.Format("02/01/2006 15:04"))
		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("B%d", i+2), revenue.Transaction)
	}

	// Tạo sheet cho phân tích
	f.NewSheet("Phân tích")
	f.SetCellValue("Phân tích", "A1", "Phân tích")
	f.SetCellValue("Phân tích", "A2", analysis)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// Ensure that the AdminService implements the IAdminService interface
var _ IAdminService = (*AdminService)(nil)
