# PESU Captive Portal — Protocol Reference

Reverse-engineered from a working reference implementation (bash script using `curl`).
This document describes only the portal's own request/response behavior — no
application-specific logic.

## Base URL

```
https://rr.pes.edu:8090
```

Overridable via the `CAPTIVE_BYPASS_PORTAL` environment variable in the reference
implementation.

## Endpoints

### Login

```
POST https://rr.pes.edu:8090/login.xml
Content-Type: application/x-www-form-urlencoded
```

Body (URL-encoded form fields):

| Field         | Value                              |
|---------------|-------------------------------------|
| `mode`        | `191` (fixed)                       |
| `username`    | portal username (SRN)               |
| `password`    | plaintext password, URL-encoded     |
| `a`            | current epoch time in milliseconds |
| `producttype` | `0` (fixed)                         |

No cookies, no CSRF token, no prior session setup — the request is stateless.

**Full response shape** (confirmed via live browser capture and curl), example on success:
```xml
<?xml version='1.0' ?>
<requestresponse>
  <status><![CDATA[LIVE]]></status>
  <message><![CDATA[You are signed in as {username}]]></message>
  <logoutmessage><![CDATA[You have successfully logged off]]></logoutmessage>
  <state><![CDATA[]]></state>
  <user><![CDATA[]]></user>
</requestresponse>
```

Example on failure (invalid credentials, confirmed via curl with throwaway creds):
```xml
<?xml version='1.0' ?>
<requestresponse>
  <status><![CDATA[LOGIN]]></status>
  <message><![CDATA[Login failed. Invalid user name/password. Please contact the administrator. ]]></message>
  <logoutmessage><![CDATA[You have successfully logged off]]></logoutmessage>
  <state><![CDATA[]]></state>
  <user><![CDATA[]]></user>
</requestresponse>
```

Fields:
- `status` — `LIVE` on success, `LOGIN` on failure (i.e. still at the login stage).
- `message` — human-readable status message; contains the failure reason on failure.
- `logoutmessage` — a static logout confirmation string, present on every response
  regardless of whether logout was requested. Not meaningful as a success/failure
  signal for login.
- `state` — empty in both observed responses; purpose unconfirmed.
- `user` — empty in both observed responses; purpose unconfirmed (possibly only
  populated in some other request mode not yet observed).

Response `Content-Type`: `text/xml`.

## Confirmed request headers (via curl `-v`)

```
POST /login.xml HTTP/1.1
Host: rr.pes.edu:8090
User-Agent: curl/8.22.0
Accept: */*
Content-Type: application/x-www-form-urlencoded
Content-Length: 66
```

No cookies or auth headers sent or required — matches the "stateless" behavior
described below.

## Server details (from curl `-v`, informational only)

- TLS: TLSv1.3, cipher `TLS_AES_256_GCM_SHA384`.
- Certificate: `CN=rr.pes.edu`, issued by Let's Encrypt, valid Aug 4 2026 – Nov 2
  2026 as of this capture. Certificate verification fails against the system
  trust store in this environment (reason unconfirmed — could be an internal/
  captive-portal-injected cert not in the system CA bundle); requests must
  disable TLS verification to succeed on this network path.
- The response passes through an internal proxy: `Via: HTTPS/1.1
  forward.http.proxy:3128`.
- Response `Connection: keep-alive`.

### Logout

```
POST https://rr.pes.edu:8090/logout.xml
Content-Type: application/x-www-form-urlencoded
```

Body:

| Field         | Value                              |
|---------------|-------------------------------------|
| `mode`        | `193` (fixed)                       |
| `username`    | portal username (SRN)               |
| `a`            | current epoch time in milliseconds |
| `producttype` | `0` (fixed)                         |

No response body is parsed by the reference implementation — the request is
fire-and-forget; any curl failure is silently ignored.

## Notes on request pacing

The reference implementation sends a logout request immediately before a login
attempt, to clear any stale session, with a short configurable delay between the
two (default 0.5s, `CAPTIVE_BYPASS_LOGOUT_LOGIN_GAP`). This is done defensively,
not because the portal is known to require it.

Login is retried up to 3 times on failure, with a delay between attempts
(default 5s, `CAPTIVE_BYPASS_RETRY_DELAY`).

## Unconfirmed / not covered here

No endpoint for checking current authentication status ("am I logged in right
now") is present anywhere in the reference implementation — only login and
logout are exercised. Any status/livecheck endpoint used elsewhere in this
project is unverified against this reference and should not be treated as
confirmed portal behavior until independently checked (e.g. via browser
DevTools while authenticated against the real portal).
