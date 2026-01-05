package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	runAddress := getDefaultRunAddress()
	accrualSystemAddress := getDefaultAccrualSystemAddress()
	databaseURI := getDefaultDatabaseURI()
	secretKey := getDefaultSecretKey()
	runMigrations := getDefaultRunMigrations()
	logLevel := getDefaultLogLevel()

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want Config
	}{
		{
			name: "Default",
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Run Address",
			args: []string{"-a", "localhost:8888"},
			want: Config{
				RunAddress:           "localhost:8888",
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Database URI",
			args: []string{"-d", "localhost:8888"},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          "localhost:8888",
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Accrual System Address",
			args: []string{"-r", "localhost:8888"},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: "localhost:8888",
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Secret Key",
			args: []string{"-s", "qwerty"},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            "qwerty",
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Log Level",
			args: []string{"-l", "warn"},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             "warn",
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Run Migrations",
			args: []string{"-m", "true"},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        true,
			},
		},

		{
			name: "Env: all settings",
			env: map[string]string{
				"RUN_ADDRESS":            "localhost:8881",
				"DATABASE_URI":           "localhost:8882",
				"ACCRUAL_SYSTEM_ADDRESS": "localhost:8883",
				"SECRET_KEY":             "qwerty",
				"LOG_LEVEL":              "warn",
				"RUN_MIGRATIONS":         "true",
			},
			want: Config{
				RunAddress:           "localhost:8881",
				DatabaseURI:          "localhost:8882",
				AccrualSystemAddress: "localhost:8883",
				SecretKey:            "qwerty",
				LogLevel:             "warn",
				RunMigrations:        true,
			},
		},
		{
			name: "Env: Run Address",
			env: map[string]string{
				"RUN_ADDRESS": "localhost:8881",
			},
			want: Config{
				RunAddress:           "localhost:8881",
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Env: Database URI",
			env: map[string]string{
				"DATABASE_URI": "localhost:8881",
			},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          "localhost:8881",
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Env: Database URI",
			env: map[string]string{
				"ACCRUAL_SYSTEM_ADDRESS": "localhost:8881",
			},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: "localhost:8881",
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Env: Secret Key",
			env: map[string]string{
				"SECRET_KEY": "qwerty",
			},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            "qwerty",
				LogLevel:             logLevel,
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Env: Log Level",
			env: map[string]string{
				"LOG_LEVEL": "warn",
			},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             "warn",
				RunMigrations:        runMigrations,
			},
		},
		{
			name: "Env: Run Migrations",
			env: map[string]string{
				"RUN_MIGRATIONS": "true",
			},
			want: Config{
				RunAddress:           runAddress,
				DatabaseURI:          databaseURI,
				AccrualSystemAddress: accrualSystemAddress,
				SecretKey:            secretKey,
				LogLevel:             logLevel,
				RunMigrations:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения для теста
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			// Сначала применяем настройки из флагов
			cfg, err := InitConfigWithArgs(tt.args)
			assert.NoError(t, err)

			// Затем применяем настройки из переменных окружения
			cfg, err = initConfigWithEnv(cfg)
			assert.NoError(t, err)

			assert.Equal(t, *cfg, tt.want)
		})
	}
}
