package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
	"strings"
)

var db *sql.DB

func main() {
	db, _ = sql.Open("mysql", "root:qishiwang5@tcp(127.0.0.1:3306)/dashboard_slo?charset=utf8")
	changeExcel("搜索上云方案-华东机群梳理-2023M5.xlsx")

}

func getDataFromdb(fullName string) int {
	name := fullName[3:]
	var totID int
	sql := fmt.Sprintf("select totId from tots where fullName like '%s'", "%"+name)
	rows := db.QueryRow(sql)
	rows.Scan(&totID)
	return totID
}

func getDataFromZYdb(fullName string) int {
	name := fullName[3:]
	var totID int
	sql := fmt.Sprintf("select totId from tots where fullName like '%s'", "%网页-自研云,"+name)
	rows := db.QueryRow(sql)
	rows.Scan(&totID)
	return totID
}

func changeExcel(path string) error {
	f, _ := excelize.OpenFile(path)
	rows, _ := f.GetRows("模块梳理表2023.5")
	for i, row := range rows {
		for j, colCell := range row {
			if i == 0 {
				break
			}
			if j == 0 {
				val := strings.TrimSpace(colCell)
				totIdZY := getDataFromZYdb(val)
				if totIdZY == 0 {
					fmt.Println(val)
				} else {
					f.SetCellValue("模块梳理表2023.5", "B"+fmt.Sprintf("%d", i+1), totIdZY)
				}
			}
		}
	}
	f.Save()
	fmt.Println("finish")
	return nil
}
