package cfg

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// defaultBaseUrl        = "http://127.0.0.1"
	defaultListenPort     = 4110
	defaultMusicBrainzUrl = "https://musicbrainz.org"
)

const (
	// BASE_URL_ENV                  = "KOITO_BASE_URL"
	DATABASE_URL_ENV              = "KOITO_DATABASE_URL"
	BIND_ADDR_ENV                 = "KOITO_BIND_ADDR"
	LISTEN_PORT_ENV               = "KOITO_LISTEN_PORT"
	ENABLE_STRUCTURED_LOGGING_ENV = "KOITO_ENABLE_STRUCTURED_LOGGING"
	ENABLE_FULL_IMAGE_CACHE_ENV   = "KOITO_ENABLE_FULL_IMAGE_CACHE"
	LOG_LEVEL_ENV                 = "KOITO_LOG_LEVEL"
	MUSICBRAINZ_URL_ENV           = "KOITO_MUSICBRAINZ_URL"
	MUSICBRAINZ_RATE_LIMIT_ENV    = "KOITO_MUSICBRAINZ_RATE_LIMIT"
	ENABLE_LBZ_RELAY_ENV          = "KOITO_ENABLE_LBZ_RELAY"
	LBZ_RELAY_URL_ENV             = "KOITO_LBZ_RELAY_URL"
	LBZ_RELAY_TOKEN_ENV           = "KOITO_LBZ_RELAY_TOKEN"
	CONFIG_DIR_ENV                = "KOITO_CONFIG_DIR"
	DEFAULT_USERNAME_ENV          = "KOITO_DEFAULT_USERNAME"
	DEFAULT_PASSWORD_ENV          = "KOITO_DEFAULT_PASSWORD"
	DISABLE_DEEZER_ENV            = "KOITO_DISABLE_DEEZER"
	DISABLE_COVER_ART_ARCHIVE_ENV = "KOITO_DISABLE_COVER_ART_ARCHIVE"
	DISABLE_MUSICBRAINZ_ENV       = "KOITO_DISABLE_MUSICBRAINZ"
	SKIP_IMPORT_ENV               = "KOITO_SKIP_IMPORT"
	ALLOWED_HOSTS_ENV             = "KOITO_ALLOWED_HOSTS"
	CORS_ORIGINS_ENV              = "KOITO_CORS_ALLOWED_ORIGINS"
	DISABLE_RATE_LIMIT_ENV        = "KOITO_DISABLE_RATE_LIMIT"
	THROTTLE_IMPORTS_MS           = "KOITO_THROTTLE_IMPORTS_MS"
	IMPORT_BEFORE_UNIX_ENV        = "KOITO_IMPORT_BEFORE_UNIX"
	IMPORT_AFTER_UNIX_ENV         = "KOITO_IMPORT_AFTER_UNIX"
)

type config struct {
	bindAddr   string
	listenPort int
	configDir  string
	// baseUrl              string
	databaseUrl          string
	musicBrainzUrl       string
	musicBrainzRateLimit int
	logLevel             int
	structuredLogging    bool
	enableFullImageCache bool
	lbzRelayEnabled      bool
	lbzRelayUrl          string
	lbzRelayToken        string
	defaultPw            string
	defaultUsername      string
	disableDeezer        bool
	disableCAA           bool
	disableMusicBrainz   bool
	skipImport           bool
	allowedHosts         []string
	allowAllHosts        bool
	allowedOrigins       []string
	disableRateLimit     bool
	importThrottleMs     int
	userAgent            string
	importBefore         time.Time
	importAfter          time.Time
}

var (
	globalConfig *config
	once         sync.Once
	lock         sync.RWMutex
)

// Initialize initializes the global configuration using the provided getenv function.
func Load(getenv func(string) string, version string) error {
	var err error
	once.Do(func() {
		globalConfig, err = loadConfig(getenv, version)
	})
	return err
}

// loadConfig loads the configuration from environment variables.
func loadConfig(getenv func(string) string, version string) (*config, error) {
	cfg := new(config)

	cfg.databaseUrl = getenv(DATABASE_URL_ENV)
	if cfg.databaseUrl == "" {
		return nil, errors.New("required parameter " + DATABASE_URL_ENV + " not provided")
	}
	cfg.bindAddr = getenv(BIND_ADDR_ENV)
	var err error
	cfg.listenPort, err = strconv.Atoi(getenv(LISTEN_PORT_ENV))
	if err != nil {
		cfg.listenPort = defaultListenPort
	}
	cfg.musicBrainzRateLimit, err = strconv.Atoi(getenv(MUSICBRAINZ_RATE_LIMIT_ENV))
	if err != nil {
		cfg.musicBrainzRateLimit = 1
	}
	cfg.musicBrainzUrl = getenv(MUSICBRAINZ_URL_ENV)
	if cfg.musicBrainzUrl == "" {
		cfg.musicBrainzUrl = defaultMusicBrainzUrl
	}
	if parseBool(getenv(ENABLE_LBZ_RELAY_ENV)) {
		cfg.lbzRelayEnabled = true
		cfg.lbzRelayToken = getenv(LBZ_RELAY_TOKEN_ENV)
		cfg.lbzRelayUrl = getenv(LBZ_RELAY_URL_ENV)
	}

	beforeutx, _ := strconv.ParseInt(getenv(IMPORT_BEFORE_UNIX_ENV), 10, 64)
	afterutx, _ := strconv.ParseInt(getenv(IMPORT_AFTER_UNIX_ENV), 10, 64)

	if beforeutx > 0 {
		cfg.importBefore = time.Unix(beforeutx, 0)
	}
	if afterutx > 0 {
		cfg.importAfter = time.Unix(afterutx, 0)
	}

	cfg.importThrottleMs, _ = strconv.Atoi(getenv(THROTTLE_IMPORTS_MS))

	cfg.disableRateLimit = parseBool(getenv(DISABLE_RATE_LIMIT_ENV))

	cfg.structuredLogging = parseBool(getenv(ENABLE_STRUCTURED_LOGGING_ENV))

	cfg.enableFullImageCache = parseBool(getenv(ENABLE_FULL_IMAGE_CACHE_ENV))
	cfg.disableDeezer = parseBool(getenv(DISABLE_DEEZER_ENV))
	cfg.disableCAA = parseBool(getenv(DISABLE_COVER_ART_ARCHIVE_ENV))
	cfg.disableMusicBrainz = parseBool(getenv(DISABLE_MUSICBRAINZ_ENV))
	cfg.skipImport = parseBool(getenv(SKIP_IMPORT_ENV))

	cfg.userAgent = fmt.Sprintf("Koito %s (contact@koito.io)", version)

	if getenv(DEFAULT_USERNAME_ENV) == "" {
		cfg.defaultUsername = "admin"
	} else {
		cfg.defaultUsername = getenv(DEFAULT_USERNAME_ENV)
	}
	if getenv(DEFAULT_PASSWORD_ENV) == "" {
		cfg.defaultPw = "changeme"
	} else {
		cfg.defaultPw = getenv(DEFAULT_PASSWORD_ENV)
	}

	cfg.configDir = getenv(CONFIG_DIR_ENV)
	if cfg.configDir == "" {
		cfg.configDir = "/etc/koito"
	}

	rawHosts := getenv(ALLOWED_HOSTS_ENV)
	cfg.allowedHosts = strings.Split(rawHosts, ",")
	cfg.allowAllHosts = cfg.allowedHosts[0] == "*"

	rawCors := getenv(CORS_ORIGINS_ENV)
	cfg.allowedOrigins = strings.Split(rawCors, ",")

	switch strings.ToLower(getenv(LOG_LEVEL_ENV)) {
	case "debug":
		cfg.logLevel = 0
	case "warn":
		cfg.logLevel = 2
	case "error":
		cfg.logLevel = 3
	case "fatal":
		cfg.logLevel = 4
	default:
		cfg.logLevel = 1
	}
	return cfg, nil
}

func parseBool(s string) bool {
	if strings.ToLower(s) == "true" {
		return true
	} else {
		return false
	}
}

// Global accessors for configuration values

func UserAgent() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.userAgent
}

func ListenAddr() string {
	lock.RLock()
	defer lock.RUnlock()
	return fmt.Sprintf("%s:%d", globalConfig.bindAddr, globalConfig.listenPort)
}

func ConfigDir() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.configDir
}

// func BaseUrl() string {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	return globalConfig.baseUrl
// }

func DatabaseUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.databaseUrl
}

func MusicBrainzUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.musicBrainzUrl
}

func MusicBrainzRateLimit() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.musicBrainzRateLimit
}

func LogLevel() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.logLevel
}

func StructuredLogging() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.structuredLogging
}

func LbzRelayEnabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayEnabled
}

func LbzRelayUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayUrl
}

func LbzRelayToken() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayToken
}

func DefaultPassword() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.defaultPw
}

func DefaultUsername() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.defaultUsername
}

func FullImageCacheEnabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.enableFullImageCache
}

func DeezerDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableDeezer
}

func CoverArtArchiveDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableCAA
}

func MusicBrainzDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableMusicBrainz
}

func SkipImport() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.skipImport
}

func AllowedHosts() []string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowedHosts
}

func AllowAllHosts() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowAllHosts
}

func AllowedOrigins() []string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowedOrigins
}

func RateLimitDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableRateLimit
}

func ThrottleImportMs() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.importThrottleMs
}

// returns the before, after times, in that order
func ImportWindow() (time.Time, time.Time) {
	return globalConfig.importBefore, globalConfig.importAfter
}
