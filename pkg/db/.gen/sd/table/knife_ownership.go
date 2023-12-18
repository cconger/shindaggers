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

var KnifeOwnership = newKnifeOwnershipTable("sd", "knife_ownership", "")

type knifeOwnershipTable struct {
	mysql.Table

	// Columns
	UserID        mysql.ColumnInteger
	KnifeID       mysql.ColumnInteger
	TransType     mysql.ColumnString
	TransactedAt  mysql.ColumnTimestamp
	InstanceID    mysql.ColumnInteger
	WasSubscriber mysql.ColumnBool
	IsVerified    mysql.ColumnBool
	EditionID     mysql.ColumnInteger

	AllColumns     mysql.ColumnList
	MutableColumns mysql.ColumnList
}

type KnifeOwnershipTable struct {
	knifeOwnershipTable

	NEW knifeOwnershipTable
}

// AS creates new KnifeOwnershipTable with assigned alias
func (a KnifeOwnershipTable) AS(alias string) *KnifeOwnershipTable {
	return newKnifeOwnershipTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new KnifeOwnershipTable with assigned schema name
func (a KnifeOwnershipTable) FromSchema(schemaName string) *KnifeOwnershipTable {
	return newKnifeOwnershipTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new KnifeOwnershipTable with assigned table prefix
func (a KnifeOwnershipTable) WithPrefix(prefix string) *KnifeOwnershipTable {
	return newKnifeOwnershipTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new KnifeOwnershipTable with assigned table suffix
func (a KnifeOwnershipTable) WithSuffix(suffix string) *KnifeOwnershipTable {
	return newKnifeOwnershipTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newKnifeOwnershipTable(schemaName, tableName, alias string) *KnifeOwnershipTable {
	return &KnifeOwnershipTable{
		knifeOwnershipTable: newKnifeOwnershipTableImpl(schemaName, tableName, alias),
		NEW:                 newKnifeOwnershipTableImpl("", "new", ""),
	}
}

func newKnifeOwnershipTableImpl(schemaName, tableName, alias string) knifeOwnershipTable {
	var (
		UserIDColumn        = mysql.IntegerColumn("user_id")
		KnifeIDColumn       = mysql.IntegerColumn("knife_id")
		TransTypeColumn     = mysql.StringColumn("trans_type")
		TransactedAtColumn  = mysql.TimestampColumn("transacted_at")
		InstanceIDColumn    = mysql.IntegerColumn("instance_id")
		WasSubscriberColumn = mysql.BoolColumn("was_subscriber")
		IsVerifiedColumn    = mysql.BoolColumn("is_verified")
		EditionIDColumn     = mysql.IntegerColumn("edition_id")
		allColumns          = mysql.ColumnList{UserIDColumn, KnifeIDColumn, TransTypeColumn, TransactedAtColumn, InstanceIDColumn, WasSubscriberColumn, IsVerifiedColumn, EditionIDColumn}
		mutableColumns      = mysql.ColumnList{UserIDColumn, KnifeIDColumn, TransTypeColumn, TransactedAtColumn, WasSubscriberColumn, IsVerifiedColumn, EditionIDColumn}
	)

	return knifeOwnershipTable{
		Table: mysql.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID:        UserIDColumn,
		KnifeID:       KnifeIDColumn,
		TransType:     TransTypeColumn,
		TransactedAt:  TransactedAtColumn,
		InstanceID:    InstanceIDColumn,
		WasSubscriber: WasSubscriberColumn,
		IsVerified:    IsVerifiedColumn,
		EditionID:     EditionIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
