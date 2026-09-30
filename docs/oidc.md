# OpenID Connect (OIDC) login

Packwiz Web can let users sign in through any standards-compliant OpenID Connect
identity provider (IdP), such as Keycloak, Authentik or Google. Several
providers can be configured at the same time, and one user can link several
accounts.

Local username/password login is never disabled. The built-in `admin` account
is never created, linked or changed through OIDC, so it stays available as a
break-glass login if an IdP is down or misconfigured.

## Server setup

Two environment variables matter:

| var                 | description                                                                                                                                                                          |
|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `PWW_PUBLIC_URL`    | External URL of the app, for example `https://packwiz.example.com` (no trailing path). Used to build the redirect URI. OIDC login does not work until this is set.                     |
| `PWW_SESSION_SECRET`| Must be set to your own long random value. Provider client secrets are encrypted with a key derived from it, so providers cannot be saved while the built-in default is in use.       |

> [!WARNING]
> Changing `PWW_SESSION_SECRET` makes stored client secrets unreadable. Affected
> providers are marked **broken** in the admin page and are hidden from the login
> page until you open each one and enter the client secret again. Local login keeps
> working the whole time.

Providers are managed in the UI, not through environment variables: sign in as an
admin and open **OIDC Providers** in the navigation (`/admin/oidc`).

## Redirect URI

Register this redirect (callback) URI with your IdP, one per provider:

```
<PWW_PUBLIC_URL>/api/v1/auth/oidc/<slug>/callback
```

`<slug>` is the short name you choose for the provider in the admin form (lowercase
letters, digits and hyphens). For example
`https://packwiz.example.com/api/v1/auth/oidc/keycloak/callback`. The admin form
shows the exact value with a copy button.

The app uses the authorization code flow with PKCE, and a confidential client
(client id and client secret).

## Provider fields

| field                | notes                                                                                                                     |
|----------------------|---------------------------------------------------------------------------------------------------------------------------|
| Display name         | Shown on the login button.                                                                                                |
| Slug                 | Used in the URLs. Changing it changes the redirect URI you must register.                                                 |
| Issuer URL           | Must be `https` (plain `http` is accepted only for `localhost`). The app reads `<issuer>/.well-known/openid-configuration`. |
| Client ID / secret   | From your IdP. The secret is write-only: it is never shown again, and leaving it blank on edit keeps the stored value.    |
| Scopes               | Must include `openid`. Default: `openid profile email`.                                                                   |
| Enabled              | Disabled providers do not appear on the login page.                                                                       |
| Create users         | Off by default. When on, a person the IdP authenticates who has no account yet gets a new non-admin user. Needs a verified email. |
| Link by email        | Off by default. When on, a sign-in whose email is **verified** by the IdP is linked to the existing user with that email. |

Use **Test discovery** to check the issuer URL before saving.

### Who can sign in

The app has no domain or group allow-list. If **Create users** is on, anyone your
IdP authenticates can get an account. Restricting who may authenticate (by
realm, group, allowed users, or an OAuth consent-screen setting) is the IdP
administrator's job.

## How sign-in decides

For every successful IdP login, in order:

1. An identity already linked for this provider and subject (`sub`) signs that user in.
2. Otherwise, if **Link by email** is on and the IdP says the email is verified,
   the matching local user is linked and signed in.
3. Otherwise, if **Create users** is on and the IdP says the email is verified,
   a new user is created and signed in.
4. Otherwise the login is rejected.

Identities are matched on provider plus `sub`, never on email alone. Deactivated
users are rejected. On any failure the user is sent back to the login page with
a generic message; the IdP's own error text is never shown.

## Users

Users manage their own linked accounts on the profile page: link another
provider, or unlink one. Unlinking is refused when it would leave the account
with no way to sign in, so users created through OIDC must set a password (same
page) or link a second account first. Admins can also remove a linked account from
the user's page.

Deleting a provider removes the identities linked to it. The delete
confirmation shows how many users would be left with no way to sign in.

## Provider examples

### Keycloak

1. In your realm, create a client with **Client authentication** on and the
   **Standard flow** enabled.
2. Set the valid redirect URI to the URI above.
3. Issuer URL: `https://<keycloak-host>/realms/<realm>`.
4. Copy the client secret from the **Credentials** tab.

### Authentik

1. Create an **OAuth2/OpenID Provider** with client type *Confidential*.
2. Add the redirect URI above as a strict redirect URI.
3. Create an application that uses the provider, and set its access policies to
   control who may sign in.
4. Issuer URL: `https://<authentik-host>/application/o/<application-slug>`.
5. Authentik's default email scope mapping reports `email_verified` as false. If you
   use **Create users** or **Link by email**, customize the mapping so it returns
   `true` for emails you trust.

### Google

1. In Google Cloud Console, create an OAuth client of type **Web application**.
2. Add the redirect URI above as an authorized redirect URI.
3. Issuer URL: `https://accounts.google.com`.

Google returns `email_verified`, so **Link by email** is safe to use. With
**Create users** on, any Google account can sign in unless you restrict the
OAuth app to your organization (set the consent screen to *Internal*).

## Deployment notes

- Rate limiting is not done by the app. Put it at your reverse proxy or load
  balancer if you need it.
- The OIDC round trip uses a short-lived `SameSite=Lax` cookie, because the
  session cookie is `SameSite=Strict` in production and is not sent when the IdP
  redirects back. Two tabs starting a login at once can make the older tab fail
  with the generic error; retry.
- Outside `development` mode the state cookie is marked `Secure`, so serve the
  app over HTTPS (for example terminate TLS at the reverse proxy) and set
  `PWW_PUBLIC_URL` to the external `https` URL.
