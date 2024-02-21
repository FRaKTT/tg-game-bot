package config

import "os"

// GetAPIKey возвращает ключ для телеграм-бота
func GetAPIKey() string {
	return os.Getenv("TGBOTAPIKEY")
}
