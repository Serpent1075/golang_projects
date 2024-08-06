package config

import "github.com/spf13/viper"

type Config struct {
	DBPort            string `mapstructure:"DBPORT"`
	DBUrl             string `mapstructure:"DB_URL"`
	JWTSecretKey      string `mapstructure:"JWT_SECRET_KEY"`
	ListeningPort     string `mapstructure:"LISTENING_PORT"`
	AuthSvcTargetPort string `mapstructure:"AUTH_SVC_TARGET_PORT"`
	AuthSvcUrl        string `mapstructure:"AUTH_SVC_URL"`
	ProductSvcUrl     string `mapstructure:"PRODUCT_SVC_URL"`
	OrderSvcUrl       string `mapstructure:"ORDER_SVC_URL"`
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath("./pkg/config/envs")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	return
}
