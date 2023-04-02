# Shindaggers

Helping a joke go too far.

A streamer friend of mine built a glorious knife pulling simulator for their community.  Knives are created and
uploaded by their community on the discord and then viewers can participate in the knife pulling as a community
using channel points on their stream.

This simple site is attempting to allow people to browse their inventories and eventually maybe even engage
in trade.

`cmd/importer` is for uploading the `bladechain` document that the streamer maintains into the db schema.
`cmd/server` is the websever

Still working on the presentation of the inventory.  But this has the schema and scaffolding to be able to
present the views.  Now I just need to return better presentation.
