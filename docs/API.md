# BareRTC Web API

BareRTC provides some web API endpoints over HTTP to support better integration with your website.

Authentication to the API endpoints is gated by the `AdminAPIKey` value in your settings.toml file.

For better integration with your website, the chat server exposes some data via JSON APIs ready for cross-origin ajax requests. In your settings.toml set the `CORSHosts` to your list of website domains, such as "https://www.example.com", "http://localhost:8080" or so on.

Current API endpoints include:

## GET /api/statistics

Returns basic info about the count and usernames of connected chatters:

```json
{
    "UserCount": 1,
    "Usernames": ["admin"]
}
```

## POST /api/authentication

This endpoint can provide JWT authentication token signing on behalf of your website. The [Chatbot](Chatbot.md) program calls this endpoint for authentication.

Post your desired JWT claims to the endpoint to customize your user and it will return a signed token for the WebSocket protocol.

```json
{
    "APIKey": "from settings.toml",
    "Claims": {
        "sub": "username",
        "nick": "Display Name",
        "op": false,
        "img": "/static/photos/avatar.png",
        "url": "/users/username",
        "emoji": "🤖",
        "gender": "m"
    }
}
```

The return schema looks like:

```json
{
    "OK": true,
    "Error": "error string, omitted if none",
    "JWT": "jwt token string"
}
```

## POST /api/shutdown

Shut down (and hopefully, reboot) the chat server. It is equivalent to the `/shutdown` operator command issued in chat, but callable from your web application. It is also used as part of deadlock detection on the BareBot chatbot.

It requires the AdminAPIKey to post:

```json
{
    "APIKey": "from settings.toml"
}
```

The return schema looks like:

```json
{
    "OK": true,
    "Error": "error string, omitted if none"
}
```

The HTTP server will respond OK, and then shut down a couple of seconds later, attempting to send a ChatServer broadcast first (as in the `/shutdown` command). If the chat server is deadlocked, this broadcast won't go out but the program will still exit.

It is up to your process supervisor to automatically restart BareRTC when it exits.

## POST /api/blocklist

Your server may pre-cache the user's blocklist for them **before** they
enter the chat room. Your site will use the `AdminAPIKey` parameter that
matches the setting in BareRTC's settings.toml (by default, a random UUID
is generated the first time).

The request payload coming from your site will be an application/json
post body like:

```json
{
    "APIKey": "from your settings.toml",
    "Username": "soandso",
    "Blocklist": [ "usernames", "that", "they", "block" ],
}
```

The server holds onto these in memory and when that user enters the chat
room (**JWT authentication only**) the front-end page will embed their
cached blocklist. When they connect to the WebSocket server, they send a
`blocklist` message to push their blocklist to the server -- it is
basically a bulk `mute` action that mutes all these users pre-emptively:
the user will not see their chat messages and the muted users can not see
the user's webcam when they broadcast later, the same as a regular `mute`
action.

The JSON response to this endpoint may look like:

```json
{
    "OK": true,
    "Error": "if error, or this key is omitted if OK"
}
```

## POST /api/block/now

Your website can tell BareRTC to put a block between users "now." For
example, if a user on your main website adds a block on another user,
and one or both of them are presently logged into the chat room, BareRTC
can begin enforcing the block immediately so both users will disappear
from each other's view and no longer get one another's messages.

The request body payload looks like:

```json
{
    "APIKey": "from your settings.toml",
    "Usernames": [ "alice", "bob" ]
}
```

The pair of usernames should be the two who are blocking each other, in
any order. This will put in a two-way block between those chatters.

If you provide more than two usernames, the block is put between every
combination of usernames given.

The JSON response to this endpoint may look like:

```json
{
    "OK": true,
    "Error": "if error, or this key is omitted if OK"
}
```

## POST /api/disconnect/now

Your website can tell BareRTC to remove a user from the chat room "now"
in case that user is presently online in the chat.

The request body payload looks like:

```json
{
    "APIKey": "from your settings.toml",
    "Usernames": [ "alice" ],
    "Message": "a custom ChatServer message to send them, optional",
    "Kick": false,
}
```

The `Message` parameter, if provided, will be sent to that user as a
ChatServer error before they are removed from the room. You can use this
to provide them context as to why they are being kicked. For example:
"You have been logged out of chat because you deactivated your profile on
the main website."

The `Kick` boolean is whether the removal should manifest to other users
in chat as a "kick" (sending a presence message of "has been kicked from
the room!"). By default (false), BareRTC will tell the user to disconnect
and it will manifest as a regular "has left the room" event to other online
chatters.

The JSON response to this endpoint may look like:

```json
{
    "OK": true,
    "Removed": 1,
    "Error": "if error, or this key is omitted if OK"
}
```

The "Removed" field is the count of users actually removed from chat; a zero
means the user was not presently online.

# Ajax Endpoints (User API)

## POST /api/profile

Fetch profile information from your main website about a user in the
chat room.

Note: this API request is done by the BareRTC chat front-end page, as an
ajax request for a current logged-in user. It backs the profile card pop-up
widget in the chat room when a user clicks on another user's profile.

The request body payload looks like:

```json
{
    "JWTToken": "the caller's chat jwt token",
    "Username": "soandso"
}
```

The JWT token is the current chat user's token. This API only works when
your BareRTC config requires the use of JWT tokens for authorization.

BareRTC will translate the request into the
["Profile Webhook"](Webhooks.md#Profile%20Webhook) to fetch the target
user's profile from your website.

The response JSON given to the chat page from /api/profile looks like:

```json
{
    "OK": true,
    "Error": "only on error messages",
    "ProfileFields": [
        {
            "Name": "Age",
            "Value": "30yo"
        },
        {
            "Name": "Gender",
            "Value": "Man"
        },
        ...
    ]
}
```

## POST /api/message/history

Load prior history in a Direct Message conversation with another party.

Note: this API request is done by the BareRTC chat front-end page, as an
ajax request for a current logged-in user.

The request body payload looks like:

```json
{
    "JWTToken": "the caller's chat jwt token",
    "Username": "soandso",
    "BeforeID": 1234
}
```

The JWT token is the current chat user's token. This API only works when
your BareRTC config requires the use of JWT tokens for authorization.

The "BeforeID" parameter is for pagination, and is optional. By default,
the first page of recent messages are returned. To get the next page, provide
the "BeforeID" which matches the MessageID of the oldest message from that
page. The endpoint will return messages having an ID before this ID.

The response JSON given to the chat page from /api/profile looks like:

```javascript
{
    "OK": true,
    "Error": "only on error messages",
    "Messages": [
        {
            // Standard BareRTC Messages.
            "username": "soandso",
            "message": "hello!",
            "msgID": 1234,
            "timestamp": "2024-01-01 11:22:33"
        }
    ],
    "Remaining": 12
}
```

The "Remaining" integer in the result shows how many older messages still
remain to be retrieved, and tells the front-end page that it can request
another page.

## POST /api/message/clear

Clear stored direct message history for a user.

This endpoint can be called by the user themself (using JWT token authorization),
or by your website (using your admin APIKey) so your site can also clear chat
history remotely (e.g., for when your user deleted their account).

The request body payload looks like:

```javascript
{
    // when called from the BareRTC frontend for the current user
    "JWTToken": "the caller's chat jwt token",

    // when called from your website
    "APIKey": "your AdminAPIKey from settings.toml",
    "Username": "soandso"
}
```

The response JSON given to the chat page looks like:

```javascript
{
    "OK": true,
    "Error": "only on error messages",
    "MessagesErased": 42
}
```
