package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/repository"
)

// func TestGeneratePack(t *testing.T) {
// 	currDir, err := os.Getwd()
// 	if err != nil {
// 		panic("ERROR GETTING CUR DIR")
// 	}
// 	core.LoadEnvironment(filepath.Dir(currDir))

// 	db, err := db.NewDB()
// 	if err != nil {
// 		t.Error("DB could not be connected")
// 	}

// 	packService := NewPackService(repository.NewPackRepo(db))
// 	packConfigId := uint64(8)
// 	pack, err := packService.GeneratePacks(&packConfigId, )
// 	if err != nil {
// 		fmt.Println("Pack: ", pack)
// 		t.Error(err)
// 	}
// }

func TestNumber(t *testing.T) {
	i := 5
	p := &i
	*p = i + 5
	fmt.Println(*p)
}

func TestRand(t *testing.T) {
	for i := 0; i < 10000; i++ {
		if rand.Intn(2-1)+1 == 2 {
			t.Error("Failed, it was 2")
		}
	}
}

func TestGeneratePacks(t *testing.T) {
	id := uint64(1)
	vendorId := "123"
	title := "Best Pack Ever"
	qty := 100
	itemQty := 5
	currentStock := 1000

	packConfig := model.PackConfig{
		ID:           &id,
		VendorID:     &vendorId,
		Title:        &title,
		Qty:          &qty,
		ItemQty:      &itemQty,
		CurrentStock: &currentStock,
	}

	packItemConfigs := []*model.PackItemConfig{}
	for i := 1; i < 10; i++ {
		id := uint64(i)
		packConfigId := uint64(1)
		itemId := uint64(i)
		qty := 50
		packItemConfigs = append(packItemConfigs, &model.PackItemConfig{
			ID:           &id,
			PackConfigID: &packConfigId,
			ItemID:       &itemId,
			Qty:          &qty,
		})
	}

	result, _ := GeneratePackItemIds(context.TODO(), packItemConfigs, &packConfig)
	fmt.Println("RESULT: ", result)
}

// func TestUploadPacks(t *testing.T) {
// 	packs := make([]*model.PackFact, 10)
// 	for i := 0; i < 10; i++ {
// 		active := true
// 		packConfigId := uint64(11)
// 		packFact := model.PackFact{
// 			Active:       &active,
// 			PackConfigID: &packConfigId,
// 		}
// 		packs[i] = &packFact
// 	}
// 	currDir, err := os.Getwd()
// 	if err != nil {
// 		panic("ERROR GETTING CUR DIR")
// 	}
// 	core.LoadEnvironment(filepath.Dir(currDir))

// 	db, err := db.NewDB()
// 	if err != nil {
// 		t.Error("DB could not be connected")
// 	}
// 	defer db.Close()
// 	packRepo := repository.NewPackRepo(db)
// 	packIds, err := packRepo.UploadPacks(packs)
// 	if err != nil {
// 		fmt.Println(err)
// 		t.Error(err)
// 	}
// 	fmt.Println(packIds)
// }

func TestPackGeneration(t *testing.T) {
	currDir, err := os.Getwd()
	if err != nil {
		panic("ERROR GETTING CUR DIR")
	}
	core.LoadLocalEnvironment(filepath.Dir(currDir))

	db, err := db.NewDB("", "", "", "", "", "")
	if err != nil {
		t.Error("DB could not be connected")
	}
	defer db.Close()

	packService := NewPackService(repository.NewPackRepo(db, nil))
	id := uint64(15)
	err = packService.GeneratePacks(context.TODO(), id, "", nil)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}
}
