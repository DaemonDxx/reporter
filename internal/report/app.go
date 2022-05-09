package report

import (
	"fmt"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var logger = log.New()

type App struct {
	config   *Config
	cron     *cron.Cron
	reporter Reporter
	sender   Sender
}

func NewApp(config *Config) *App {
	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	reporter := NewExcelReporter(db, config)
	sender := NewMailSender(&config.Subscribers, &config.MailClient)
	app := App{
		cron:     cron.New(cron.WithLocation(time.UTC)),
		sender:   sender,
		reporter: reporter,
		config:   config,
	}
	return &app
}

func (a *App) Run() {
	logger.Info("Инициализация приложения")
	_, err := a.cron.AddFunc(a.config.Schedule, func() {
		logger.Info("Формирование нового отчета")
		err := a.updateReport()
		if err != nil {
			log.Errorf("Не удалось сформировать и отправить отчет")
		} else {
			log.Info("Отчет направлен")
		}
	})
	if err != nil {
		log.Fatalln(err)
	}
	logger.Info("Приложение запущено")
	a.cron.Start()
}

func (a *App) updateReport() error {
	report, err := a.reporter.Build()
	if err != nil {
		return err
	}
	currentDate := time.Now()
	msg := fmt.Sprintf("Актуальный отчет по температурному фактору на %s", timeToString(&currentDate))
	err = a.sender.Send(msg, report)
	return err
}

func timeToString(time *time.Time) string {
	return fmt.Sprintf("%d.%d.%d", time.Day(), time.Month(), time.Year())
}
