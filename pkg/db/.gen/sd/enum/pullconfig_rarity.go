//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/mysql"

var PullconfigRarity = &struct {
	Uncommon  mysql.StringExpression
	Common    mysql.StringExpression
	Rare      mysql.StringExpression
	UltraRare mysql.StringExpression
	SuperRare mysql.StringExpression
}{
	Uncommon:  mysql.NewEnumValue("Uncommon"),
	Common:    mysql.NewEnumValue("Common"),
	Rare:      mysql.NewEnumValue("Rare"),
	UltraRare: mysql.NewEnumValue("Ultra Rare"),
	SuperRare: mysql.NewEnumValue("Super Rare"),
}
