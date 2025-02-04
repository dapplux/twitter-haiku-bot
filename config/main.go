package config

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string `split_words:"true"`
}

type Twitter struct {
	APIBearer            string `split_words:"true"`
	APIKey               string `split_words:"true"`
	APISecret            string `split_words:"true"`
	APIAccessToken       string `split_words:"true"`
	APIAccessTokenSecret string `split_words:"true"`
}

type HuggingFace struct {
	APIKey string `split_words:"true"`
}

type Config struct {
	DB          DB
	Twitter     Twitter
	HuggingFace HuggingFace
}

type Source interface {
	Load() (Config, error)
}

func Load(source Source) (Config, error) {
	return source.Load()
}

func AutoLoad() (Config, error) {
	envSrc := EnvSource{
		Prefix: "",
	}
	cfg, err := Load(envSrc)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
