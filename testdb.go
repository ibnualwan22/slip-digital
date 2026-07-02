package main

import (
	"fmt"
	"log"

	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
)

func main() {
	cfg := config.LoadConfig()
	var activities []domain.MasterActivity
	err := cfg.DB.Find(&activities).Error
	if err != nil {
		log.Fatalf("Error finding activities: %v", err)
	}
	fmt.Printf("Found %d activities\n", len(activities))
}
