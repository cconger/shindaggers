//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/mysql"
)

var Equipped = newEquippedTable("sd", "equipped", "")

type equippedTable struct {
	mysql.Table

	// Columns
	UserID     mysql.ColumnInteger
	InstanceID mysql.ColumnInteger
	EquippedAt mysql.ColumnTimestamp

	AllColumns     mysql.ColumnList
	MutableColumns mysql.ColumnList
}

type EquippedTable struct {
	equippedTable

	NEW equippedTable
}

// AS creates new EquippedTable with assigned alias
func (a EquippedTable) AS(alias string) *EquippedTable {
	return newEquippedTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new EquippedTable with assigned schema name
func (a EquippedTable) FromSchema(schemaName string) *EquippedTable {
	return newEquippedTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new EquippedTable with assigned table prefix
func (a EquippedTable) WithPrefix(prefix string) *EquippedTable {
	return newEquippedTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new EquippedTable with assigned table suffix
func (a EquippedTable) WithSuffix(suffix string) *EquippedTable {
	return newEquippedTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newEquippedTable(schemaName, tableName, alias string) *EquippedTable {
	return &EquippedTable{
		equippedTable: newEquippedTableImpl(schemaName, tableName, alias),
		NEW:           newEquippedTableImpl("", "new", ""),
	}
}

func newEquippedTableImpl(schemaName, tableName, alias string) equippedTable {
	var (
		UserIDColumn     = mysql.IntegerColumn("user_id")
		InstanceIDColumn = mysql.IntegerColumn("instance_id")
		EquippedAtColumn = mysql.TimestampColumn("equipped_at")
		allColumns       = mysql.ColumnList{UserIDColumn, InstanceIDColumn, EquippedAtColumn}
		mutableColumns   = mysql.ColumnList{InstanceIDColumn, EquippedAtColumn}
	)

	return equippedTable{
		Table: mysql.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID:     UserIDColumn,
		InstanceID: InstanceIDColumn,
		EquippedAt: EquippedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
