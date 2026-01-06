package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	SecretKey     string `env:"SECRET_KEY"`
	LogLevel      string `env:"LOG_LEVEL"`
	RunMigrations bool   `env:"RUN_MIGRATIONS"`

	MinPasswordLength int `env:"MIN_PASSWORD_LENGTH"`
	MaxPasswordLength int `env:"MAX_PASSWORD_LENGTH"`

	MinLoginLength int `env:"MIN_LOGIN_LENGTH"`
	MaxLoginLength int `env:"MAX_LOGIN_LENGTH"`
}

// PrintConfig выводит настройки в консоль
func (c *Config) PrintConfig() {
	fmt.Printf("Run address: %s\n", c.RunAddress)
	fmt.Printf("Database URI: %s\n", c.DatabaseURI)
	fmt.Printf("Accrual System Address: %s\n", c.AccrualSystemAddress)
	fmt.Printf("Secret Key: %s\n", c.SecretKey)
	fmt.Printf("Log Level: %s\n", c.LogLevel)
	fmt.Printf("Run Migrations: %t\n", c.RunMigrations)
	fmt.Printf("MinPasswordLength: %d\n", c.MinPasswordLength)
	fmt.Printf("MaxPasswordLength: %d\n", c.MaxPasswordLength)
	fmt.Printf("MinLoginLength: %d\n", c.MinLoginLength)
	fmt.Printf("MaxLoginLength: %d\n", c.MaxLoginLength)
}

// InitConfig возвращает настройки и ошибку если парсинг аргументов не удался
func InitConfig() (*Config, error) {
	// Инициализируем настройки из флагов, также будут заданы значения по умолчанию
	config, err := InitConfigWithArgs(os.Args[1:])

	if err != nil {
		return nil, err
	}

	// Инициализируем настройки из переменных окружения если таковые указаны
	config, err = initConfigWithEnv(config)

	if err != nil {
		return nil, err
	}

	var errs []error

	if config.RunAddress == "" {
		errs = append(errs, errors.New("run address is required"))
	}

	if config.DatabaseURI == "" {
		errs = append(errs, errors.New("database URI is required"))
	}

	if config.AccrualSystemAddress == "" {
		errs = append(errs, errors.New("accrual system address is required"))
	}

	if config.SecretKey == "" {
		errs = append(errs, errors.New("secret key is required"))
	}

	if config.MinPasswordLength < 3 || config.MinPasswordLength > 255 {
		errs = append(errs, errors.New("min password length must be between 3 and 255"))
	}

	if config.MaxPasswordLength < 3 || config.MaxPasswordLength > 255 {
		errs = append(errs, errors.New("max password length must be between 3 and 255"))
	}

	if config.MaxPasswordLength < config.MinPasswordLength {
		errs = append(errs, errors.New("max password length must be greater or equal to min password length"))
	}

	if config.MinLoginLength < 3 || config.MinLoginLength > 255 {
		errs = append(errs, errors.New("min login length must be between 3 and 255"))
	}

	if config.MaxLoginLength < 3 || config.MaxLoginLength > 255 {
		errs = append(errs, errors.New("max login length must be between 3 and 255"))
	}

	if config.MaxLoginLength < config.MinLoginLength {
		errs = append(errs, errors.New("max login length must be greater or equal to min password length"))
	}

	return config, errors.Join(errs...)
}

// initConfigWithEnv получение настроек из переменных окружения
func initConfigWithEnv(config *Config) (*Config, error) {
	if runAddress, present := os.LookupEnv("RUN_ADDRESS"); present {
		config.RunAddress = runAddress
	}

	if databaseURI, present := os.LookupEnv("DATABASE_URI"); present {
		config.DatabaseURI = databaseURI
	}

	if accrualSystemAddress, present := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); present {
		config.AccrualSystemAddress = accrualSystemAddress
	}

	if secretKey, present := os.LookupEnv("SECRET_KEY"); present {
		config.SecretKey = secretKey
	}

	if logLevel, present := os.LookupEnv("LOG_LEVEL"); present {
		config.LogLevel = logLevel
	}

	if runMigrations, present := os.LookupEnv("RUN_MIGRATIONS"); present {
		config.RunMigrations, _ = strconv.ParseBool(runMigrations)
	}

	if err := parseIntFromEnv(config, "MIN_PASSWORD_LENGTH",
		func(c *Config, v int) { c.MinPasswordLength = v }); err != nil {
		return nil, err
	}

	if err := parseIntFromEnv(config, "MAX_PASSWORD_LENGTH",
		func(c *Config, v int) { c.MaxPasswordLength = v }); err != nil {
		return nil, err
	}

	if err := parseIntFromEnv(config, "MIN_LOGIN_LENGTH",
		func(c *Config, v int) { c.MinLoginLength = v }); err != nil {
		return nil, err
	}

	if err := parseIntFromEnv(config, "MAX_LOGIN_LENGTH",
		func(c *Config, v int) { c.MaxLoginLength = v }); err != nil {
		return nil, err
	}

	return config, nil
}

// InitConfigWithArgs инициализация с переданными аргументами
func InitConfigWithArgs(args []string) (*Config, error) {
	var config Config

	flagSet := flag.NewFlagSet("main", flag.ContinueOnError)

	flagSet.StringVarP(&config.RunAddress, "run-address", "a", getDefaultRunAddress(), "run address")
	flagSet.StringVarP(&config.DatabaseURI, "database-uri", "d", getDefaultDatabaseURI(), "database URI")
	flagSet.StringVarP(&config.AccrualSystemAddress, "accrual-system-address", "r", getDefaultAccrualSystemAddress(), "accrual system address")

	flagSet.StringVarP(&config.LogLevel, "log-level", "l", getDefaultLogLevel(), "log level")
	flagSet.StringVarP(&config.SecretKey, "secret-key", "s", getDefaultSecretKey(), "secret key")
	flagSet.BoolVarP(&config.RunMigrations, "run-migrations", "m", getDefaultRunMigrations(), "run migrations")

	flagSet.IntVarP(&config.MinPasswordLength, "min-password-length", "o", getDefaultMinPasswordLength(), "min password length")
	flagSet.IntVarP(&config.MaxPasswordLength, "max-password-length", "p", getDefaultMaxPasswordLength(), "max password length")

	flagSet.IntVarP(&config.MinLoginLength, "min-login-length", "q", getDefaultMinLoginLength(), "min login length")
	flagSet.IntVarP(&config.MaxLoginLength, "max-login-length", "t", getDefaultMaxLoginLength(), "max login length")

	err := flagSet.Parse(args)

	if err != nil {
		return nil, err
	}

	// Возвращаем адрес переменной config
	return &config, nil
}

// parseIntFromEnv парсит int-значение из переменной окружения и устанавливает его в поле конфигурации
func parseIntFromEnv(config *Config, envKey string, setter func(*Config, int)) error {
	if value, present := os.LookupEnv(envKey); present {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid env %s %s", envKey, value)
		}
		setter(config, intValue)
	}
	return nil
}

// getDefaultRunAddress Стандартные настройки подключения к БД
func getDefaultRunAddress() string {
	return "http://127.0.0.1:8080"
}

// getDefaultDatabaseURI Стандартные настройки подключения к БД
func getDefaultAccrualSystemAddress() string {
	return "http://127.0.0.1:8080"
}

// getDefaultDatabaseURI Стандартные настройки подключения к БД
func getDefaultDatabaseURI() string {
	return ""
}

// getDefaultSecretKey секретный ключ по умолчанию
func getDefaultSecretKey() string {
	return "gophermat_secret_key"
}

// getDefaultLogLevel уровень логирования по умолчанию
func getDefaultLogLevel() string {
	return "info"
}

// getDefaultRunMigrations запуск миграций по умолчанию
func getDefaultRunMigrations() bool {
	return false
}

// getDefaultMinPasswordLength минимальная длина пароля
func getDefaultMinPasswordLength() int {
	return 3
}

// getDefaultMaxPasswordLength максимальная длина пароля
func getDefaultMaxPasswordLength() int {
	return 255
}

// getDefaultMinLoginLength минимальная длина логина
func getDefaultMinLoginLength() int {
	return 3
}

// getDefaultMaxLoginLength максимальная длина логина
func getDefaultMaxLoginLength() int {
	return 255
}
