package database

import (
	"gorm.io/gorm"
	"irbtc_streamer/database/table"
	"log"
)

func MigrateDatabaseTables(db *gorm.DB) error {
	log.Println("⚙️ Running AutoMigrate for all tables...")

	models := []interface{}{
		&table.PriceHistoryTable{},
		&table.VolumeHistoryTable{},
		&table.PriceChangesTable{},
		&table.HTXSymbolTable{},
		// add more models(table) here as needed
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("❌ AutoMigrate failed for model %T: %v\n", model, err)
			return err
		}
		log.Printf("✅ AutoMigrate completed successfully for model %T.\n", model)
	}

	return nil
}
