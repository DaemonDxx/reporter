package report

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const ConfigPath = "/etc/reporter"

type Config struct {
	TempDir     string         `yaml:"tempDir"`
	DBPath      string         `yaml:"db"`
	Template    TemplateConfig `yaml:"template"`
	MailClient  MailConfig     `yaml:"mailClient"`
	Subscribers []string       `yaml:"subscribers"`
	Schedule    string         `yaml:"schedule"`
}

type TemplateConfig struct {
	Filepath  string `yaml:"filepath"`
	Sheetname string `yaml:"sheetname"`
}

type MailConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

func (c *Config) Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(ConfigPath)
	viper.AddConfigPath("./")
	viper.AddConfigPath("./configs/")
	viper.OnConfigChange(func(in fsnotify.Event) {
		newCfg := &Config{}
		err := viper.Unmarshal(newCfg)
		if err != nil {
			log.Fatalln(err)
		}
		c.Subscribers = newCfg.Subscribers
		log.WithField("module", "config").Info("Конфигурация обновлена")
	})
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(c)
	if err != nil {
		return err
	}
	return nil
}
