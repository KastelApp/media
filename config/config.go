package config

type Config struct {
	CdnUrl string `json:"cdnUrl"`
	Secure bool   `json:"secure"`
}

func GetProtocol(secure bool) string {
	if secure {
		return "https"
	}
	return "http"
}