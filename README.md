# Shindaggers

Helping a joke go too far.

A streamer friend of mine built a glorious knife pulling simulator for their community.  Knives are created and
uploaded by their community on the discord and then viewers can participate in the knife pulling as a community
using channel points on their stream.

`cmd/importer` is for uploading the `bladechain` document that the streamer maintains into the db schema.
`cmd/server` is the websever


### Warning

This code is not the best.  What I like about it is that I do most things manually.  

Things I want to improve if there is any traction:
 - Rethink templating pretty completely
 - Caching of responses both on server and with Cache Control headers
 - Transactions on dbs

