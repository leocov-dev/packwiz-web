# Running behind a reverse proxy

If packwiz-web sits behind a reverse proxy or load balancer (Traefik, nginx,
Caddy, a cloud load balancer, ...), tell it which proxies to trust. Otherwise it
sees every request as coming from the proxy.

## What happens when it is not set

`PWW_TRUSTED_PROXIES` is empty by default. The app then ignores the
`X-Forwarded-For` header and uses the address of the direct connection. Behind a
proxy that is the proxy's own IP, so the **Audit** page shows the same IP for
every user.

## Configure

Set `PWW_TRUSTED_PROXIES` to the IP address or CIDR range of your proxy, as seen
by packwiz-web. Separate multiple entries with commas.

```yaml
environment:
  - PWW_TRUSTED_PROXIES=172.18.0.0/16
```

```
PWW_TRUSTED_PROXIES=10.0.0.5,192.168.1.0/24
```

Finding the proxy address:

- Docker: the subnet of the network shared by the proxy and the app
  (`docker network inspect <network>`).
- Otherwise: the address in the app's connection logs, or the proxy host's IP.

Only list proxies you control. Any listed address can set the client IP by
sending `X-Forwarded-For`, so do not trust a range that untrusted hosts can
reach.

## The proxy must forward the client IP

Most proxies add `X-Forwarded-For` by default. If yours does not, enable it.

## Multiple proxies

With a chain such as CDN → reverse proxy → packwiz-web, each hop must trust the
one before it, and packwiz-web must trust the last hop. If an earlier hop is not
trusted, it replaces the real IP with its own.

## Check it works

Open the Audit page after making a request from a known address. The IP shown
should be yours, not the proxy's.
