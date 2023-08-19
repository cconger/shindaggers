# Shindaggers

Helping a joke go too far.

A streamer friend of mine built a glorious knife pulling simulator for their community.  Knives are created and
uploaded by their community on the discord and then viewers can participate in the knife pulling as a community
using channel points on their stream.

`cmd/importer` is for uploading the `bladechain` document that the streamer maintains into the db schema.
`cmd/server` is the websever

## Developing

To get the API running you can run
```
go run ./cmd/server -nodb
```

This will start the server in a special mode that uses a mock db and the templates will be reloaded every
time.  You can load it at http://localhost:8080

If you want to use real data, you unfortuantely need several secrets for the twitch client and to access the database set through env vars:

`CLOUDFLARE_SECRET`
`CLOUDFLARE_CLIENT_ID`
`STORAGE_ENDPOINT`
`TWITCH_CLIENT_ID`
`TWITCH_SECRET`
`DSN`


## Web application

If you just want to work on the presentation you can run the webapp in standalone mode see [client/README.md](./client/README.md)

## Stuff in progress

 - [x] Rethink templating pretty completely (move to js front end)
 - [x] Server side pulling
   - [x] Animations for pull in webapp
 - [x] Admin Pages Moved to JS App
 - [x] Allow logged in subscribers to upload knives direct to site pending approval
 - [x] Allow logged in users to "equip" a knife
 - [ ] FIght leaderboards and stats
   - [ ] Event page for knife fights
 - [ ] Live "Latest"
 - [ ] Fix embedding, titles and metadata returned by server
 - [ ] Local Dev database that isn't garbage

###  Exploration Ideas

 - [ ] Allow users to trade knives?
 - [ ] Allow users to turn knives into matierals to craft other knives?
