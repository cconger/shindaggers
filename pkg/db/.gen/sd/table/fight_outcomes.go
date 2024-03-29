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

var FightOutcomes = newFightOutcomesTable("sd", "fight_outcomes", "")

type fightOutcomesTable struct {
	mysql.Table

	// Columns
	FightID       mysql.ColumnInteger
	UserID        mysql.ColumnInteger
	CollectableID mysql.ColumnInteger
	Outcome       mysql.ColumnString
	EventID       mysql.ColumnInteger

	AllColumns     mysql.ColumnList
	MutableColumns mysql.ColumnList
}

type FightOutcomesTable struct {
	fightOutcomesTable

	NEW fightOutcomesTable
}

// AS creates new FightOutcomesTable with assigned alias
func (a FightOutcomesTable) AS(alias string) *FightOutcomesTable {
	return newFightOutcomesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new FightOutcomesTable with assigned schema name
func (a FightOutcomesTable) FromSchema(schemaName string) *FightOutcomesTable {
	return newFightOutcomesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new FightOutcomesTable with assigned table prefix
func (a FightOutcomesTable) WithPrefix(prefix string) *FightOutcomesTable {
	return newFightOutcomesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new FightOutcomesTable with assigned table suffix
func (a FightOutcomesTable) WithSuffix(suffix string) *FightOutcomesTable {
	return newFightOutcomesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newFightOutcomesTable(schemaName, tableName, alias string) *FightOutcomesTable {
	return &FightOutcomesTable{
		fightOutcomesTable: newFightOutcomesTableImpl(schemaName, tableName, alias),
		NEW:                newFightOutcomesTableImpl("", "new", ""),
	}
}

func newFightOutcomesTableImpl(schemaName, tableName, alias string) fightOutcomesTable {
	var (
		FightIDColumn       = mysql.IntegerColumn("fight_id")
		UserIDColumn        = mysql.IntegerColumn("user_id")
		CollectableIDColumn = mysql.IntegerColumn("collectable_id")
		OutcomeColumn       = mysql.StringColumn("outcome")
		EventIDColumn       = mysql.IntegerColumn("event_id")
		allColumns          = mysql.ColumnList{FightIDColumn, UserIDColumn, CollectableIDColumn, OutcomeColumn, EventIDColumn}
		mutableColumns      = mysql.ColumnList{CollectableIDColumn, OutcomeColumn, EventIDColumn}
	)

	return fightOutcomesTable{
		Table: mysql.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		FightID:       FightIDColumn,
		UserID:        UserIDColumn,
		CollectableID: CollectableIDColumn,
		Outcome:       OutcomeColumn,
		EventID:       EventIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
