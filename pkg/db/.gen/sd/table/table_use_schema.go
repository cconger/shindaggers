//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

// UseSchema sets a new schema name for all generated table SQL builder types. It is recommended to invoke
// this method only once at the beginning of the program.
func UseSchema(schema string) {
	Editions = Editions.FromSchema(schema)
	Equipped = Equipped.FromSchema(schema)
	Events = Events.FromSchema(schema)
	FightOutcomes = FightOutcomes.FromSchema(schema)
	Fights = Fights.FromSchema(schema)
	ImageUploads = ImageUploads.FromSchema(schema)
	KnifeOwnership = KnifeOwnership.FromSchema(schema)
	Knives = Knives.FromSchema(schema)
	Pullconfig = Pullconfig.FromSchema(schema)
	UserAuth = UserAuth.FromSchema(schema)
	Users = Users.FromSchema(schema)
}
