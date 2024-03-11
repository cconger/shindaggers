//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var UserEquipCollectableInstance = newUserEquipCollectableInstanceTable("public", "user_equip_collectable_instance", "")

type userEquipCollectableInstanceTable struct {
	postgres.Table

	// Columns
	UserID     postgres.ColumnInteger
	InstanceID postgres.ColumnInteger
	EquippedAt postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type UserEquipCollectableInstanceTable struct {
	userEquipCollectableInstanceTable

	EXCLUDED userEquipCollectableInstanceTable
}

// AS creates new UserEquipCollectableInstanceTable with assigned alias
func (a UserEquipCollectableInstanceTable) AS(alias string) *UserEquipCollectableInstanceTable {
	return newUserEquipCollectableInstanceTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserEquipCollectableInstanceTable with assigned schema name
func (a UserEquipCollectableInstanceTable) FromSchema(schemaName string) *UserEquipCollectableInstanceTable {
	return newUserEquipCollectableInstanceTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UserEquipCollectableInstanceTable with assigned table prefix
func (a UserEquipCollectableInstanceTable) WithPrefix(prefix string) *UserEquipCollectableInstanceTable {
	return newUserEquipCollectableInstanceTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UserEquipCollectableInstanceTable with assigned table suffix
func (a UserEquipCollectableInstanceTable) WithSuffix(suffix string) *UserEquipCollectableInstanceTable {
	return newUserEquipCollectableInstanceTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUserEquipCollectableInstanceTable(schemaName, tableName, alias string) *UserEquipCollectableInstanceTable {
	return &UserEquipCollectableInstanceTable{
		userEquipCollectableInstanceTable: newUserEquipCollectableInstanceTableImpl(schemaName, tableName, alias),
		EXCLUDED:                          newUserEquipCollectableInstanceTableImpl("", "excluded", ""),
	}
}

func newUserEquipCollectableInstanceTableImpl(schemaName, tableName, alias string) userEquipCollectableInstanceTable {
	var (
		UserIDColumn     = postgres.IntegerColumn("user_id")
		InstanceIDColumn = postgres.IntegerColumn("instance_id")
		EquippedAtColumn = postgres.TimestampColumn("equipped_at")
		allColumns       = postgres.ColumnList{UserIDColumn, InstanceIDColumn, EquippedAtColumn}
		mutableColumns   = postgres.ColumnList{InstanceIDColumn, EquippedAtColumn}
	)

	return userEquipCollectableInstanceTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID:     UserIDColumn,
		InstanceID: InstanceIDColumn,
		EquippedAt: EquippedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}