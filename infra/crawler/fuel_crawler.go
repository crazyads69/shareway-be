package crawler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"shareway/infra/db/migration"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

// IFuelCrawler defines the interface for fuel price crawling operations
type IFuelCrawler interface {
	UpdateFuelPrices() error
	FetchFuelPrices() ([]migration.FuelPrice, error)
	SaveFuelPrices(prices []migration.FuelPrice) error
}

// FuelCrawler implements the IFuelCrawler interface
type FuelCrawler struct {
	db *gorm.DB
}

// NewFuelCrawler creates a new FuelCrawler instance
func NewFuelCrawler(db *gorm.DB) IFuelCrawler {
	return &FuelCrawler{db: db}
}

// UpdateFuelPrices updates the fuel prices in the database
func (fc *FuelCrawler) UpdateFuelPrices() error {
	prices, err := fc.FetchFuelPrices()
	if err != nil {
		return err
	}

	return fc.SaveFuelPrices(prices)
}

// FetchFuelPrices fetches the latest fuel prices from the website
func (fc *FuelCrawler) FetchFuelPrices() ([]migration.FuelPrice, error) {
	url := "https://vnexpress.net/chu-de/gia-xang-dau-3026"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var prices []migration.FuelPrice

	doc.Find("table").Each(func(i int, tableHTML *goquery.Selection) {
		tableHTML.Find("tr").Each(func(rowIndex int, rowHTML *goquery.Selection) {
			var rowData []string
			rowHTML.Find("td").Each(func(cellIndex int, cellHTML *goquery.Selection) {
				rowData = append(rowData, strings.TrimSpace(cellHTML.Text()))
			})

			if len(rowData) >= 2 && rowData[0] != "Mặt hàng" {
				fuelType := rowData[0]
				priceStr := strings.ReplaceAll(rowData[1], ".", "")
				price, err := strconv.ParseFloat(priceStr, 64)
				if err != nil {
					log.Printf("Failed to parse price for %s: %v\n", fuelType, err)
					return
				}

				prices = append(prices, migration.FuelPrice{
					FuelType: fuelType,
					Price:    price,
				})
			}
		})
	})

	return prices, nil
}

// SaveFuelPrices saves or updates the fuel prices in the database
func (fc *FuelCrawler) SaveFuelPrices(prices []migration.FuelPrice) error {
	return fc.db.Transaction(func(tx *gorm.DB) error {
		for _, price := range prices {
			var existingPrice migration.FuelPrice
			result := tx.Where("fuel_type = ?", price.FuelType).First(&existingPrice)

			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					if err := tx.Create(&price).Error; err != nil {
						return err
					}
				} else {
					return result.Error
				}
			} else {
				existingPrice.Price = price.Price
				if err := tx.Save(&existingPrice).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
