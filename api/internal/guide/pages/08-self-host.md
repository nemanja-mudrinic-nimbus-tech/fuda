# Self-host

fuda is one container. People log in with a GitHub App, and fuda reads each Board with that
person's own token. fuda keeps no server token and no shared password. Access is decided by GitHub.

## The GitHub App

Create one app for fuda (Settings → Developer settings → GitHub Apps):

- **Enable Device Flow:** on. People log in with a code, so the app needs no callback URL and no
  client secret.
- **Expire user authorization tokens:** off. fuda keeps no refresh token.
- **Webhook:** off.
- **Repository permissions:** Contents read, Pull requests read, Metadata read.

Install the app on the organisation or user that owns the `fuda-` repositories, for those repositories only. fuda then sees only those. Find the install page at `github.com/apps/<slug>/installations/new`.

To find the slug, open [github.com/settings/apps](https://github.com/settings/apps) as the person who created the App (for an organisation's App, its Settings → Developer settings → GitHub Apps). Click the App and read its public link: `https://github.com/apps/<slug>`. The slug is the last part.

```sh
docker run -p 8080:8080 -v fuda-data:/data \
  -e FUDA_GITHUB_CLIENT_ID=<client id> \
  -e FUDA_GITHUB_APP_SLUG=<app slug> \
  fuda
```

Keep the `/data` volume. fuda makes its login cookie key on first start and saves it there as
`cookie.key`. Without the volume every restart makes a new key and logs everyone out.

Serve fuda over HTTPS. Login cookies are `Secure` except on `localhost`, so a plain-HTTP address
such as `http://192.168.1.5:8080` has no working login.

## Settings

| Variable | Default | Meaning |
|---|---|---|
| `FUDA_HOSTS` | `github` | `github`, `azure`, `local`, or `github,azure` for both hosts at once. The old name `FUDA_SOURCE` still works |
| `FUDA_GITHUB_CLIENT_ID` | | The GitHub App's client id |
| `FUDA_GITHUB_APP_SLUG` | | The GitHub App's slug, the last part of `github.com/apps/<slug>`. Required with `github`: fuda does not start without it. It builds the install link |
| `FUDA_BASE_URL`, `FUDA_COOKIE_SECRET` | | Only for `azure`: the public address (login callback) and the key that encrypts its login cookie, at least 32 characters. Not used for GitHub |
| `FUDA_AZURE_CLIENT_ID`, `FUDA_AZURE_CLIENT_SECRET` | | The Microsoft Entra ID app's client id and secret (when `FUDA_HOSTS` includes `azure`) |
| `FUDA_AZURE_TENANT` | `organizations` | The Entra tenant that may log in. Use your tenant id for a single-tenant app |
| `FUDA_TITLE` | repository name | Shown in the header of the `local` Board |
| `FUDA_WATCH_MAIN` | `false` | Also read `main` for the "in prod" badge |
| `FUDA_SYNC_COOLDOWN` | `30s` | Minimum time between manual syncs and pull-request re-reads |
| `FUDA_CACHE_DIR` | `/data` in the image | Where the last read copy and the GitHub login cookie key are kept. Must persist |

`/healthz` never asks for a login. The `local` host has no login and a server shows it read-only (only
the desktop app writes to local folders), so keep such a server on a private network or behind your
proxy's authentication.

## How people see Boards

- Not logged in: a Board page sends you to the login page. For GitHub it shows a code to type on github.com. Then you come back.
- After login the Board button in the top bar lists the `fuda-` repositories the app is installed
  on and you can read. `/` opens the only Board, or lists them.
- App not installed: login still works, but the list is empty. fuda checks this after login and shows a banner with the install link. People who cannot install the App can send the link to an owner of the organisation.
- No access: the page says so. Check that the app is installed on the repository and that your
  account can read it.
- The token lives in an encrypted cookie. fuda stores no user secret.
- Moving from an old GitHub setup: fuda ignores `FUDA_GITHUB_CLIENT_SECRET`, `FUDA_COOKIE_SECRET` and `FUDA_BASE_URL` and logs one warning. Remove them.

## Updates

The browser asks for changes about every 5 seconds. fuda checks the head commit with a conditional
request, which costs no rate limit when nothing changed, and reads files only when the head moved.
There are no webhooks.
