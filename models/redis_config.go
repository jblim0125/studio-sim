package models

// RedisConfig redis config
type RedisConfig struct {
	IP          string `yaml:"ip" json:"ip"`
	Port        int    `yaml:"port" json:"port"`
	Database    string `yaml:"database" json:"database"`
	SubDatabase string `yaml:"subDatabase" json:"subDatabase"`
	ID          string `yaml:"id" json:"id"`
	PW          string `yaml:"pw" json:"pw"`
}
