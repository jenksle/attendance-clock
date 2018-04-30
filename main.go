package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

//Config sctructure
type Config struct {
	ClockServer string `json:"clockServer"`
	ClockUser   string `json:"clockUser"`
	ClockPwd    string `json:"clockPwd"`
	ClockDb     string `json:"clockDb"`
	Port        int    `json:"port"`
}

//Department struct
type department struct {
	DepartmentID   int
	DepartmentName string
}

type employee struct {
	FirstName  string
	Surname    string
	EmployeeID int
}

type employeeclock struct {
	FirstName   string
	Surname     string
	EmployeeID  int
	TimeID      int
	StartDT     string
	StartTime   string
	FinishDT    string
	TimeDiff    string
	ClockInTime string
}

type deptpage struct {
	Departments []department
	ServerTime  string
}

type employeepage struct {
	Employees1     []employee
	Employees2     []employee
	DepartmentName string
	ServerTime     string
}

type clockpage struct {
	FirstName      string
	Surname        string
	EmployeeID     string
	EmployeeClocks []employeeclock
	ClockDetail    string
	InOut          string
	ClockedIn      string
	DepartmentName string
	DepartmentID   int
	ServerTime     string
}

var db *sql.DB
var config Config

func init() {

	// Load application configuration from settings file
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the mysql database and test connection
	connection := fmt.Sprintf("%s:%s@/%s",
		config.ClockUser,
		config.ClockPwd,
		config.ClockDb)

	db, err = sql.Open("mysql", connection)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

}

func main() {

	router := httprouter.New()

	router.ServeFiles("/assets/*filepath", http.Dir("./assets"))

	router.GET("/", departments)
	router.GET("/employees/:departmentid", deptemployees)
	router.GET("/employeedetails/:employeeid", employeedetails)
	router.GET("/startstop/:employeeid", startstop)
	router.GET("/time", gettime)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}

func departments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var output deptpage

	//Retreive department ID and name
	sql1 := `select department_id, department_name from tbl_department`

	rows, err := db.Query(sql1)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var d department
	output.Departments = make([]department, 0)

	for rows.Next() {

		err := rows.Scan(&d.DepartmentID, &d.DepartmentName)
		if err != nil {

			log.Fatal(err)
		}

		output.Departments = append(output.Departments, d)

	}
	fmt.Println(output.Departments)

	output.ServerTime = time.Now().Format("15 : 04 : 05")

	t, err := template.ParseFiles("templates/index.tpl")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, output)
	if err != nil {
		log.Fatal(err)
	}

}

func deptemployees(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var output employeepage

	//check department id is an integer - if not throw an error
	_, err := strconv.Atoi(ps.ByName("departmentid"))
	if err != nil {
		log.Fatal("Department ID is not numeric")
	}

	departmentID, _ := strconv.Atoi(ps.ByName("departmentid"))

	sqldept := `SELECT department_name 
			FROM tbl_department
			WHERE department_id = ?`

	err = db.QueryRow(sqldept, departmentID).Scan(&output.DepartmentName)
	if err != nil {
		log.Fatal(err)
	}

	sql2 := `SELECT first_name, surname, employee_id
				FROM tbl_employee e
				WHERE is_active = 'Y'
				  AND department_id = ?
				ORDER BY surname ASC`

	rows, err := db.Query(sql2, departmentID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var e employee
	output.Employees1 = make([]employee, 0)
	output.Employees2 = make([]employee, 0)

	count := 0

	for rows.Next() {

		err := rows.Scan(&e.FirstName, &e.Surname, &e.EmployeeID)
		if err != nil {
			log.Fatal(err)
		}

		if count < 10 {
			output.Employees1 = append(output.Employees1, e)
		} else {
			output.Employees2 = append(output.Employees2, e)
		}
		count++
	}

	output.ServerTime = time.Now().Format("15 : 04 : 05")

	t, err := template.ParseFiles("templates/employees.tpl")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, output)
	if err != nil {
		log.Fatal(err)
	}

}

func employeedetails(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//check employee id is an integer - if not throw an error
	//convert to an int variable and store as an int employee id

	_, err := strconv.Atoi(ps.ByName("employeeid"))
	if err != nil {
		log.Fatal("Employee ID is not numeric")
	}

	employeeID, _ := strconv.Atoi(ps.ByName("employeeid"))

	var output clockpage

	sqldept := `SELECT d.department_name, d.department_id 
				FROM tbl_employee e
					INNER JOIN tbl_department d ON d.department_id = e.department_id
				WHERE employee_id = ?`

	err = db.QueryRow(sqldept, employeeID).Scan(&output.DepartmentName, &output.DepartmentID)
	if err != nil {
		log.Fatal(err)
	}

	var c employeeclock

	sql3 := `SELECT first_name, surname, employee_id
				FROM tbl_employee
				WHERE employee_id = ?`

	err = db.QueryRow(sql3, employeeID).Scan(&output.FirstName, &output.Surname, &output.EmployeeID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("employee details obtained")

	sql4 := `SELECT e.employee_id, t.time_id, DATE_FORMAT(t.start_dt, '%W, %D %M'), DATE_FORMAT(t.start_dt, '%T'), COALESCE(DATE_FORMAT(t.finish_dt, '%T'), '-'), COALESCE(TIMEDIFF(t.finish_dt, t.start_dt), '-')
				FROM tbl_employee e
					INNER JOIN tbl_time t ON t.employee_id=e.employee_id
				WHERE e.employee_id = ?
				ORDER BY t.time_id DESC
				LIMIT 5`

	rows, err := db.Query(sql4, employeeID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	output.EmployeeClocks = make([]employeeclock, 0)

	for rows.Next() {

		err := rows.Scan(&c.EmployeeID, &c.TimeID, &c.StartDT, &c.StartTime, &c.FinishDT, &c.TimeDiff)
		if err != nil {
			log.Fatal(err)
		}

		output.EmployeeClocks = append(output.EmployeeClocks, c)
	}
	fmt.Println("20 most recent clock events obtained")

	sql5 := `SELECT e.employee_id, t.time_id, DATE_FORMAT(t.start_dt, '%W, %D %M'), COALESCE(DATE_FORMAT(t.finish_dt, '%Y-%m-%d %T'), '-'), TIMEDIFF(NOW(), t.start_dt)
				FROM tbl_employee e
					INNER JOIN tbl_time t ON t.employee_id=e.employee_id
				WHERE e.employee_id = ?
				ORDER BY t.time_id DESC
				LIMIT 1`

	err = db.QueryRow(sql5, employeeID).Scan(&c.EmployeeID, &c.TimeID, &c.StartDT, &c.FinishDT, &c.ClockInTime)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	fmt.Println("most recent clock event obtained")

	if c.FinishDT == "-" {
		output.ClockDetail = "clocked in"
		output.InOut = "Clock out"
		output.ClockedIn = c.ClockInTime

	} else {
		output.ClockDetail = "clocked out"
		output.InOut = "Clock in"
		output.ClockedIn = "00:00"
	}

	output.ServerTime = time.Now().Format("15 : 04 : 05")

	t, err := template.ParseFiles("templates/clock.tpl")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, output)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("end of employeedetails func")
}

func startstop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//check employee id is an integer - if not throw an error
	//convert to an int variable and store as an int employee id

	_, err := strconv.Atoi(ps.ByName("employeeid"))
	if err != nil {
		log.Fatal("Employee ID is not numeric")
	}

	employeeID, _ := strconv.Atoi(ps.ByName("employeeid"))

	var c employeeclock

	fmt.Printf("\n\n%s\n", ps.ByName("employeeid"))
	fmt.Println("Start of startstop func")

	currentTime := time.Now().Format("02 Jan 06 15:04 MST")

	sql6 := `SELECT t.time_id, DATE_FORMAT(t.start_dt, '%Y-%m-%d %T'), COALESCE(DATE_FORMAT(t.finish_dt, '%Y-%m-%d %T'), '-')
			 FROM tbl_employee e
				 INNER JOIN tbl_time t ON t.employee_id=e.employee_id
			 WHERE e.employee_id = ?
			 ORDER BY t.time_id DESC
			 LIMIT 1`

	err = db.QueryRow(sql6, employeeID).Scan(&c.TimeID, &c.StartDT, &c.FinishDT)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	fmt.Println("most recent clock event obtained 2")

	if c.FinishDT == "-" {

		sql7 := `UPDATE tbl_time
				 SET finish_dt = NOW()
					WHERE time_id = ?;`

		_, err := db.Exec(sql7, c.TimeID)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Clocked off at " + currentTime)

	} else {

		fmt.Println("entered else block")

		sql8 := `INSERT INTO tbl_time (employee_id, start_dt)
				 VALUES (?, NOW())`

		fmt.Println(employeeID, sql8)
		_, err := db.Exec(sql8, employeeID)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Clocked in at " + currentTime)

	}

	http.Redirect(w, r, "/", 303)

}

func gettime(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ServerTime := time.Now().Format("15 : 04 : 05")
	fmt.Fprintf(w, ServerTime)

}
