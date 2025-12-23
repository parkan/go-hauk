package config

import (
	"os"
	"strconv"
)

type AuthMethod int

const (
	AuthPassword AuthMethod = iota
	AuthHtpasswd
	AuthLDAP
)

type VelocityUnit int

const (
	KilometersPerHour VelocityUnit = iota
	MilesPerHour
	MetersPerSecond
)

type LinkStyle int

const (
	Link4Plus4Upper LinkStyle = iota
	Link4Plus4Lower
	Link4Plus4Mixed
	LinkUUIDv4
	Link16Hex
	Link16Upper
	Link16Lower
	Link16Mixed
	Link32Hex
	Link32Upper
	Link32Lower
	Link32Mixed
)

type Config struct {
	ListenAddr string
	PublicURL  string

	// redis
	RedisAddr     string
	RedisPassword string
	RedisPrefix   string

	// auth
	AuthMethod   AuthMethod
	PasswordHash string
	HtpasswdPath string

	// ldap
	LDAPUri        string
	LDAPBaseDN     string
	LDAPBindDN     string
	LDAPBindPass   string
	LDAPUserFilter string
	LDAPStartTLS   bool

	// share limits
	MaxDuration  int
	MinInterval  float64
	MaxCachedPts int
	MaxShownPts  int

	// link generation
	LinkStyle      LinkStyle
	AllowLinkReq   bool
	ReservedLinks  map[string][]string
	ReserveWL      bool

	// map display
	MapTileURI     string
	MapAttribution string
	DefaultZoom    int
	MaxZoom        int

	// velocity
	VelocityUnit      VelocityUnit
	VelocityDataPts   int
	TrailColor        string
	OfflineTimeout    int
	RequestTimeout    int
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return def
}

func Load() *Config {
	authMethod := AuthPassword
	switch envStr("HAUK_AUTH_METHOD", "password") {
	case "htpasswd":
		authMethod = AuthHtpasswd
	case "ldap":
		authMethod = AuthLDAP
	}

	velUnit := KilometersPerHour
	switch envStr("HAUK_VELOCITY_UNIT", "km/h") {
	case "mph":
		velUnit = MilesPerHour
	case "m/s":
		velUnit = MetersPerSecond
	}

	return &Config{
		ListenAddr:     envStr("HAUK_LISTEN_ADDR", ":8080"),
		PublicURL:      envStr("HAUK_PUBLIC_URL", "http://localhost:8080/"),
		RedisAddr:      envStr("HAUK_REDIS_ADDR", "localhost:6379"),
		RedisPassword:  envStr("HAUK_REDIS_PASSWORD", ""),
		RedisPrefix:    envStr("HAUK_REDIS_PREFIX", "hauk"),
		AuthMethod:     authMethod,
		PasswordHash:   envStr("HAUK_PASSWORD_HASH", ""),
		HtpasswdPath:   envStr("HAUK_HTPASSWD_PATH", "/etc/hauk/users.htpasswd"),
		LDAPUri:        envStr("HAUK_LDAP_URI", ""),
		LDAPBaseDN:     envStr("HAUK_LDAP_BASE_DN", ""),
		LDAPBindDN:     envStr("HAUK_LDAP_BIND_DN", ""),
		LDAPBindPass:   envStr("HAUK_LDAP_BIND_PASS", ""),
		LDAPUserFilter: envStr("HAUK_LDAP_USER_FILTER", "(uid=%s)"),
		LDAPStartTLS:   envBool("HAUK_LDAP_START_TLS", false),
		MaxDuration:    envInt("HAUK_MAX_DURATION", 86400),
		MinInterval:    envFloat("HAUK_MIN_INTERVAL", 1),
		MaxCachedPts:   envInt("HAUK_MAX_CACHED_PTS", 3),
		MaxShownPts:    envInt("HAUK_MAX_SHOWN_PTS", 100),
		LinkStyle:      LinkStyle(envInt("HAUK_LINK_STYLE", 0)),
		AllowLinkReq:   envBool("HAUK_ALLOW_LINK_REQ", true),
		ReservedLinks:  make(map[string][]string),
		ReserveWL:      envBool("HAUK_RESERVE_WHITELIST", false),
		MapTileURI:     envStr("HAUK_MAP_TILE_URI", "https://tile.openstreetmap.org/{z}/{x}/{y}.png"),
		MapAttribution: envStr("HAUK_MAP_ATTRIBUTION", `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors`),
		DefaultZoom:    envInt("HAUK_DEFAULT_ZOOM", 14),
		MaxZoom:        envInt("HAUK_MAX_ZOOM", 19),
		VelocityUnit:   velUnit,
		VelocityDataPts: envInt("HAUK_VELOCITY_DATA_PTS", 2),
		TrailColor:     envStr("HAUK_TRAIL_COLOR", "#d80037"),
		OfflineTimeout: envInt("HAUK_OFFLINE_TIMEOUT", 30),
		RequestTimeout: envInt("HAUK_REQUEST_TIMEOUT", 10),
	}
}
