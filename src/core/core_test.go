package core

import (
	"fmt"
	"testing"
	"xo-packs/model"
)

func TestGenerateOdds(t *testing.T) {
	rarities := []uint64{1, 1, 1, 1, 1, 2, 2, 2, 3, 3, 4, 5, 5}
	itemIds := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

	items := []model.Item{}
	for i, rarity := range rarities {
		currRarity := rarity
		currItemId := itemIds[i]
		item := model.Item{ID: &currItemId, RarityId: &currRarity}
		items = append(items, item)
	}

	odds, err := GenerateOdds(items, 1000)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(odds)
}