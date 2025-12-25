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

```
docker run -p 8080:8080 \
  -e HAUK_AUTH_METHOD=password \
  -e HAUK_PASSWORD_HASH='$2a$10$...' \
  -e HAUK_REDIS_ADDR=redis:6379 \
  ghcr.io/parkan/go-hauk
```

## config

all config via environment variables:

| var | default | description |
|-----|---------|-------------|
| HAUK_LISTEN_ADDR | :8080 | listen address |
| HAUK_PUBLIC_URL | http://localhost:8080/ | public url for share links |
| HAUK_REDIS_ADDR | localhost:6379 | redis address (host:port or redis:// url) |
| HAUK_AUTH_METHOD | password | auth method (password, htpasswd, ldap) |
| HAUK_PASSWORD_HASH | | bcrypt hash for password auth |

see `config/config.go` for full list.

## compatibility

drop-in replacement for the php backend. works with the existing android app and web frontend.

## license

same as upstream (apache 2.0).
