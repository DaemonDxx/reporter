package report

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"time"
)

const (
	indexColl      = 1
	departmentCol  = 2
	dateCol        = 3
	valueCol       = 5
	prevValueCol   = 4
	reportFilename = "report.xlsm"
	table          = "temperature_entities"
	selectQuery    = "t.department, t.temperature as temperature, t.year, t.month, t.day, t2.temperature as Prev"
	whereQuery     = "t.year = ?"
	joinQuery      = "LEFT OUTER JOIN temperature_entities as t2 ON t2.year = ? AND t2.month = t.month AND t2.day = t.day AND t2.department = t.department"
)

type Report struct {
	Filename string
	Data     io.Reader
}

type record struct {
	Temperature float32
	Department  string
	Day         int
	Month       int
	Year        int
	Prev        float32
}

type Reporter interface {
	Build() (*Report, error)
}

type ExcelReporter struct {
	db          *gorm.DB
	path        *string
	templateCfg *TemplateConfig
	log         *log.Entry
}

func NewExcelReporter(db *gorm.DB, config *Config) Reporter {
	path := path.Join(config.TempDir, reportFilename)
	return &ExcelReporter{
		db:          db,
		path:        &path,
		templateCfg: &config.Template,
		log:         log.New().WithField("module", "builder"),
	}
}

func (r *ExcelReporter) Build() (*Report, error) {
	template, err := excelize.OpenFile(r.templateCfg.Filepath)
	if err != nil {
		r.log.Errorf("Не удалось открыть шаблон: %e", err)
		return nil, err
	}
	records, err := r.getAllRecords()
	if err != nil {
		return nil, err
	}

	fillTemplate(template, records)

	err = template.SaveAs(*r.path)
	if err != nil {
		r.log.Errorf("Не удалось сохранить отчет в файловой системе: %e", err)
		return nil, err
	}

	f, err := os.Open(*r.path)
	if err != nil {
		r.log.Errorf("Не удалось открыть сохранненый отчет в файловой системе: %e", err)
		return nil, err
	}

	return &Report{
		Filename: reportFilename,
		Data:     f,
	}, err
}

func (r *ExcelReporter) getAllRecords() ([]*record, error) {
	year := time.Now().Year()
	records := make([]*record, 0)

	rows, err := r.db.Table("temperature_entities as t").Joins(joinQuery, year-1).Select(selectQuery).Where(whereQuery, year).Rows()
	if err != nil {
		r.log.Errorf("Не удалось выполнить запрос к БД: %e", err)
		return nil, err
	}
	for rows.Next() {
		record := record{}
		prev := sql.NullFloat64{}
		err := rows.Scan(&record.Department, &record.Temperature, &record.Year, &record.Month, &record.Day, &prev)
		if err != nil {
			r.log.Errorf("Не удалось получить строку из поискового запроса: %e", err)
			continue
		}
		if !prev.Valid {
			continue
		}
		record.Prev = float32(prev.Float64)
		records = append(records, &record)
	}

	return records, nil
}

func fillTemplate(template *excelize.File, items []*record) *excelize.File {
	for i, record := range items {
		indexCoord, _ := excelize.CoordinatesToCellName(indexColl, i+2)
		departmentCoord, _ := excelize.CoordinatesToCellName(departmentCol, i+2)
		dateCoord, _ := excelize.CoordinatesToCellName(dateCol, i+2)
		valueCoord, _ := excelize.CoordinatesToCellName(valueCol, i+2)
		prevValueCoord, _ := excelize.CoordinatesToCellName(prevValueCol, i+2)

		date := time.Date(record.Year, time.Month(record.Month), record.Day, 0, 0, 0, 0, time.UTC)
		template.SetCellValue("Источник", indexCoord, i)
		template.SetCellValue("Источник", departmentCoord, record.Department)
		template.SetCellValue("Источник", dateCoord, date)
		template.SetCellValue("Источник", prevValueCoord, record.Prev)
		template.SetCellValue("Источник", valueCoord, record.Temperature)
	}
	return template
}
