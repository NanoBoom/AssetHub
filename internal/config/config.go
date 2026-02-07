package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type StorageConfig struct {
	Type  string      `mapstructure:"type"`
	S3    S3Config    `mapstructure:"s3"`
	OSS   OSSConfig   `mapstructure:"oss"`
	Local LocalConfig `mapstructure:"local"`
}

type S3Config struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Endpoint        string `mapstructure:"endpoint"`
	UsePathStyle    bool   `mapstructure:"use_path_style"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
}

type LocalConfig struct {
	BasePath string `mapstructure:"base_path"`
}

func Load(path string) (*Config, error) {
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.env", "development")
	viper.SetDefault("database.max_open_conns", 10)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	_ = viper.ReadInConfig()

	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("log.level", "LOG_LEVEL")

	// Storage 配置绑定环境变量
	viper.BindEnv("storage.type", "STORAGE_TYPE")
	viper.BindEnv("storage.s3.region", "S3_REGION")
	viper.BindEnv("storage.s3.bucket", "S3_BUCKET")
	viper.BindEnv("storage.s3.access_key_id", "S3_ACCESS_KEY_ID")
	viper.BindEnv("storage.s3.secret_access_key", "S3_SECRET_ACCESS_KEY")
	viper.BindEnv("storage.s3.endpoint", "S3_ENDPOINT")
	viper.BindEnv("storage.s3.use_path_style", "S3_USE_PATH_STYLE")

	viper.BindEnv("storage.oss.endpoint", "OSS_ENDPOINT")
	viper.BindEnv("storage.oss.bucket", "OSS_BUCKET")
	viper.BindEnv("storage.oss.access_key_id", "OSS_ACCESS_KEY_ID")
	viper.BindEnv("storage.oss.access_key_secret", "OSS_ACCESS_KEY_SECRET")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
