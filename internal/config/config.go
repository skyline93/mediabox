package config

type Config struct {
	HttpHost string `json:"http_host"`
	HttpPort int    `json:"http_port"`
	UseTLS   bool   `json:"use_tls"`
	TLSCert  string `json:"tls_cert"`
	TLSKey   string `json:"tls_key"`

	DbDriver string `json:"db_driver"`
	DbDsn    string `json:"db_dsn"`

	StoragePath string `json:"storage_path"`
}
