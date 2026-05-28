# go-hauk

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/sysHvT?referralCode=PNe-Vg)

Go port of the [Hauk](https://github.com/bilde2910/Hauk) location sharing backend.

## why

The original PHP implementation works fine but has some overhead. This port provides:

- 345x higher throughput (45k vs 133 req/s)
- 36x smaller image (16MB vs 578MB)
- 3x lower memory footprint (8MB vs 25MB idle)
- single static binary, no runtime dependencies
- redis only (no memcached)

## usage

run go-hauk and redis together with docker compose:

```
# generate a password hash
export HAUK_PASSWORD_HASH=$(htpasswd -nbBC 10 "" 'your-password' | tr -d ':\n')

docker compose up -d
```

or copy `.env.example` to `.env`, fill it in, and `docker compose up -d`.

### persistence

share data (sessions, links, locations) lives in redis with a per-share TTL
(`HAUK_MAX_DURATION`, default 24h). the bundled redis persists to a named volume
by default, so restarts and redeploys keep active shares alive. set
`REDIS_PERSIST=off` for in-memory only -- faster, but every share is lost on
restart.

### deploy on railway

use the **Deploy on Railway** button above.

## config

all config via environment variables:

| var | default | description |
|-----|---------|-------------|
| HAUK_LISTEN_ADDR | :8080 | listen address |
| HAUK_PUBLIC_URL | http://localhost:8080/ | public url for share links |
| HAUK_REDIS_ADDR | localhost:6379 | redis address (host:port or redis:// url) |
| HAUK_AUTH_METHOD | password | auth method (password, htpasswd, ldap) |
| HAUK_PASSWORD_HASH | | bcrypt hash for password auth |
| HAUK_MAX_DURATION | 86400 | max share lifetime in seconds (redis key TTL) |
| HAUK_RATE_LIMIT_AUTH | 10 | max auth requests per minute per ip |
| HAUK_RATE_LIMIT_ADOPT | 10 | max adopt requests per minute per ip |
| HAUK_TRUST_PROXY | true | trust X-Forwarded-For (set false if not behind proxy) |

see `config/config.go` for full list.

## security improvements over upstream

- adopt authorization: only share owner can adopt into groups (fixes CVE-like auth bypass in upstream)
- built-in rate limiting on auth and adopt endpoints (configurable, default 10 req/min/ip)

## disclaimer

this is a best-effort port of the upstream reference backend. the security
surface has been mitigated where practical, but no third-party audit has been
performed, and risks inherent in the upstream protocol (e.g. 6-digit group pins,
view links readable by anyone with the url) are carried forward. use at your own
discretion.

## compatibility

drop-in replacement for the php backend. works with the existing android app and web frontend.

## license

same as upstream (apache 2.0).
