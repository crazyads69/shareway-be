package crawler

import (
	"log"
	"net/http"
	"regexp"
	"shareway/infra/db/migration"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

var (
	brandRegex           = regexp.MustCompile(`(?i)Nhãn hiệu\s*:([^:;]+)`)
	commercialNameRegex  = regexp.MustCompile(`(?i)Tên thương mại:\s*([^;]+)`)
	fuelConsumptionRegex = regexp.MustCompile(`(?i)Mức tiêu thụ nhiên liệu công khai\s*:?\s*([\d,\.]+)\s*(?:[lL]ít|[lL])?\s*/\s*100\s*km`)
	modelCodeRegex       = regexp.MustCompile(`(?i)Mã [Kk]iểu [Ll]oại:?\s*([^;:]+)`)
)

type IVrCrawler interface {
	CrawlData() error
	CrawlPage(url string) ([]migration.VehicleType, error)
	UpdateOrCreateVehicles(vehicles []migration.VehicleType) error
}

type VrCrawler struct {
	db *gorm.DB
}

func NewVrCrawler(db *gorm.DB) IVrCrawler {
	return &VrCrawler{db: db}
}

func (c *VrCrawler) CrawlData() error {
	baseURL := "http://www.vr.org.vn/Pages/thong-bao.aspx?Category=22&Page="
	maxPages := 57

	for page := 1; page <= maxPages; page++ {
		url := baseURL + strconv.Itoa(page)
		vehicles, err := c.CrawlPage(url)
		if err != nil {
			log.Printf("Error crawling page %d: %v", page, err)
			continue
		}

		if err := c.UpdateOrCreateVehicles(vehicles); err != nil {
			log.Printf("Error updating database for page %d: %v", page, err)
		}
	}

	return nil
}

func (c *VrCrawler) CrawlPage(url string) ([]migration.VehicleType, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var vehicles []migration.VehicleType

	doc.Find("table.tableList tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // Skip header row
			return
		}

		content := s.Find("td").Eq(1).Text()
		vehicle := extractVehicleInfo(content)
		if vehicle != nil {
			vehicles = append(vehicles, *vehicle)
		}
	})

	return vehicles, nil
}

func (c *VrCrawler) UpdateOrCreateVehicles(vehicles []migration.VehicleType) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		for _, vehicle := range vehicles {
			var existingVehicle migration.VehicleType
			result := tx.Where("name = ?", vehicle.Name).First(&existingVehicle)
			if result.Error == gorm.ErrRecordNotFound {
				if err := tx.Create(&vehicle).Error; err != nil {
					return err
				}
			} else if result.Error == nil {
				existingVehicle.FuelConsumed = vehicle.FuelConsumed
				existingVehicle.UpdatedAt = time.Now()
				if err := tx.Save(&existingVehicle).Error; err != nil {
					return err
				}
			} else {
				return result.Error
			}
		}
		return nil
	})
}

func extractVehicleInfo(content string) *migration.VehicleType {
	// Skip entries related to cars or electric vehicles
	if strings.Contains(strings.ToLower(content), "xe hơi") ||
		strings.Contains(strings.ToLower(content), "ô tô") ||
		strings.Contains(strings.ToLower(content), "wh/km") {
		return nil
	}

	brandMatch := brandRegex.FindStringSubmatch(content)
	commercialNameMatch := commercialNameRegex.FindStringSubmatch(content)
	fuelConsumptionMatch := fuelConsumptionRegex.FindStringSubmatch(content)
	modelCodeMatch := modelCodeRegex.FindStringSubmatch(content)

	var name string
	var brand, commercialName, modelCode string

	if len(brandMatch) > 1 {
		brand = strings.TrimSpace(brandMatch[1])
	}
	if len(commercialNameMatch) > 1 {
		commercialName = strings.TrimSpace(commercialNameMatch[1])
	}
	if len(modelCodeMatch) > 1 {
		modelCode = strings.TrimSpace(modelCodeMatch[1])
	}

	// Determine the name based on available information
	if brand != "" && !isEmptyOrDash(commercialName) {
		name = brand + " " + commercialName
	} else if !isEmptyOrDash(commercialName) {
		name = commercialName
	} else if brand != "" && !isEmptyOrDash(modelCode) {
		name = brand + " " + modelCode
	} else if brand != "" {
		name = brand
	} else if !isEmptyOrDash(modelCode) {
		name = modelCode
	}

	// Remove any "---" from the name and trim spaces
	name = strings.ReplaceAll(name, "---", "")
	name = strings.ReplaceAll(name, "--", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "  ", " ")
	name = strings.ReplaceAll(name, "Nhãn hiệu", "")
	name = strings.ReplaceAll(name, "Tên thương mại", "")
	name = strings.ReplaceAll(name, "Mức tiêu thụ nhiên liệu công khai", "")
	name = strings.ReplaceAll(name, "Mã Kiểu Loại", "")
	name = strings.ReplaceAll(name, "Mã kiểu loại", "")
	name = strings.ReplaceAll(name, ":", "")
	name = strings.TrimSpace(name)

	var fuelConsumption float64
	if len(fuelConsumptionMatch) > 1 {
		fuelConsumptionStr := strings.TrimSpace(fuelConsumptionMatch[1])
		fuelConsumptionStr = strings.Replace(fuelConsumptionStr, ",", ".", -1) // Replace comma with dot
		fuelConsumption, _ = strconv.ParseFloat(fuelConsumptionStr, 64)
	}

	// If we couldn't extract either name or fuel consumption, return nil
	if name == "" || fuelConsumption == 0 {
		return nil
	}

	return &migration.VehicleType{
		Name:         name,
		FuelConsumed: fuelConsumption,
	}
}

func isEmptyOrDash(s string) bool {
	trimmed := strings.TrimSpace(s)
	return trimmed == "" || trimmed == "-" || trimmed == "/" || trimmed == "---" || trimmed == "--"
}

// Make sure the crawler implements the IVrCrawler interface
var _ IVrCrawler = (*VrCrawler)(nil)
