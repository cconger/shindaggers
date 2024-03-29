//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import "errors"

type PullconfigRarity string

const (
	PullconfigRarity_Uncommon  PullconfigRarity = "Uncommon"
	PullconfigRarity_Common    PullconfigRarity = "Common"
	PullconfigRarity_Rare      PullconfigRarity = "Rare"
	PullconfigRarity_UltraRare PullconfigRarity = "Ultra Rare"
	PullconfigRarity_SuperRare PullconfigRarity = "Super Rare"
)

func (e *PullconfigRarity) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("jet: Invalid scan value for AllTypesEnum enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "Uncommon":
		*e = PullconfigRarity_Uncommon
	case "Common":
		*e = PullconfigRarity_Common
	case "Rare":
		*e = PullconfigRarity_Rare
	case "Ultra Rare":
		*e = PullconfigRarity_UltraRare
	case "Super Rare":
		*e = PullconfigRarity_SuperRare
	default:
		return errors.New("jet: Invalid scan value '" + enumValue + "' for PullconfigRarity enum")
	}

	return nil
}

func (e PullconfigRarity) String() string {
	return string(e)
}
