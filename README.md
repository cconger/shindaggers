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


### Warning

This code is not the best.  What I like about it is that I do most things manually.  

Things I want to improve if there is any traction:
 - Rethink templating pretty completely
 - Move admininstration to a javascript webapp

 - Live frontpage
 - Be able to pull serverside
   - Show animations
