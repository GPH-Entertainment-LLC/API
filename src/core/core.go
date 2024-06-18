package core

import (
	"fmt"
	"math"
	"xo-packs/model"
)

type BasicItem struct {
	ItemId uint64
	Rarity int
	Amount int
}

func DistributeBehind(items []*BasicItem, amountRemaining int, prev *BasicItem, leastRare int, deltaMap map[int]float64) {
	if len(items) == 0 {
		return
	}

	idx := len(items) - 1
	item := items[idx]

	if item.Rarity == prev.Rarity {
		item.Amount = prev.Amount
		DistributeBehind(items[:idx], amountRemaining, item, leastRare, deltaMap)
		return
	} else if item.Rarity == leastRare {
		keep := int(max(math.Floor(float64((amountRemaining)/(idx+1))), 1))
		item.Amount += keep
		amountRemaining -= keep
		DistributeBehind(items[:idx], amountRemaining, item, leastRare, deltaMap)
		return
	} else {
		deltaPct := deltaMap[item.Rarity-leastRare]
		keep := int(max(math.Floor(float64(amountRemaining/(idx+1))*deltaPct), 1))
		item.Amount += keep
		amountRemaining -= keep
		DistributeBehind(items[:idx], amountRemaining, item, leastRare, deltaMap)
		return
	}
}

func GenerateOdds(rawItems []model.Item, totalItems int) (map[uint64]int, error) {
	oddsMap := map[uint64]int{}

	if len(rawItems) == 0 {
		return nil, nil
	}

	// build basic items, determine least rare, build rarity freq map
	basicItems := []*BasicItem{}
	leastRare := 5
	rarityFreqMap := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
	for _, item := range rawItems {
		basicItem := BasicItem{
			ItemId: *item.ID,
			Rarity: int(*item.RarityId),
			Amount: 0,
		}
		basicItems = append(basicItems, &basicItem)
		if basicItem.Rarity < leastRare {
			leastRare = basicItem.Rarity
		}
		rarityFreqMap[basicItem.Rarity] += 1
	}

	// apply item amount
	baseAmount := int(math.Floor(float64(totalItems / len(basicItems))))
	deltaMap := map[int]float64{
		1: 0.75,
		2: 0.5,
		3: .15,
		4: .10,
	}
	prev := new(BasicItem)
	for i, item := range basicItems {
		if prev.ItemId != 0 {
			if item.Rarity == prev.Rarity {
				item.Amount = prev.Amount
			} else if item.Rarity > prev.Rarity {
				delta := deltaMap[item.Rarity-leastRare]
				keep := int(max(math.Floor((float64(baseAmount) * delta)), 1))
				item.Amount += keep
				amountRemaining := baseAmount - keep
				DistributeBehind(basicItems[:i], amountRemaining, item, leastRare, deltaMap)
			}
		} else {
			item.Amount = baseAmount
		}

		prev = item
	}

	// distribute remaining padding to most common
	amountDistributed := 0
	for _, item := range basicItems {
		amountDistributed += item.Amount
	}

	remaining := totalItems - amountDistributed
	padding := int(math.Floor(float64(remaining / rarityFreqMap[leastRare])))
	for _, item := range basicItems {
		if item.Rarity != leastRare {
			break
		}
		item.Amount += padding
	}

	// distribute any remainders across items
	amountDistributed = 0
	for _, item := range basicItems {
		amountDistributed += item.Amount
	}

	remaining = totalItems - amountDistributed
	for remaining > 0 {
		for _, item := range basicItems {
			item.Amount += 1
			remaining -= 1
			if remaining <= 0 {
				break
			}
		}
	}

	// verify total items count is the same as amount distributed
	if remaining < 0 {
		fmt.Println("Critical error in generating odds")
		return nil, &ErrorResp{Message: "Critical error in generating odds"}
	}

	// build result odds list
	for _, item := range basicItems {
		oddsMap[item.ItemId] = item.Amount
	}
	return oddsMap, nil
}
