Common prefix:

https://api.shindaggers.io/v1/

Send an authorization token using header

Authorization: Bearer <token>

API is json over http

# Collectable

## /collectable POST

Create a new collectable

## /collectable/{id} GET

Get a collectable by id

## /collectable/{id} PUT

Upate a collectable by id

## /collectable/{id} DELETE

Delete a collectable by id

## /collectable/community/{id} GET

Gets a list of collectables for the given community id

# User

## /user/me GET

# Returns the logged in user

## /user/{id} GET

Gets a user by id

## /user/{id}/collection?community={community_id} GET

Gets a User's collection of collectables, optionally filtered by to the passed param of communities

# Lootbox

## /lootbox POST

Create a new lootbox

## /lootbox/{id} GET

Get a given lootbox by id

## /lootbox/{id} PUT

Update a given lootbox by id

## /lootbox/{id} DELETE

Delete a given lootbox

## /lootbox/community/{id} GET

Get a list of lootboxes for a given community id

## /lootbox/{id}/pull POST

Opens a lootbox for a user

# Community 

## /community/{id} GET

Gets a community by id.

## /comunity/{id}/latest GET

Get a paginated list of the latest pulls for a community

### TODO: A way to add or remove users as admins of your community

## /collected/{id}

Gets a collected collectable by id

## /collected/{id}/history 

Gets a paginated slice of tranactions for this item

# Oauth

## /oauth/redirect

# Webhooks

## /webhook POST

Create a new webhook for a user and community

## /webhook/{id} GET

Get the webhook

## /webhook/{id} DELETE

Delete a the webhook

## /w/{token}/pull

Do a pull for your community using the specified token

