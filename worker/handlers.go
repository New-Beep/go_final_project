package worker

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pwd/pkg/db"
	"time"

	"pwd/internal/controller"
	"pwd/internal/nextdate"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {

}

type TaskController struct {
	db *sql.DB
}

func NewTaskController(db *sql.DB) *TaskController {
	return &TaskController{db: db}
}

func (c *TaskController) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.GetTasks(c.db)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type response struct {
		Tasks []controller.Task `json:"tasks"`
	}

	resp := response{Tasks: tasks}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func (c *TaskController) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task controller.Task // экзмпляр структуры со значениями
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "ошибка десериализации json", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if task.Title == "" {
		http.Error(w, "Поле title обязательно", http.StatusBadRequest)
		return
	}

	// проверяем дату
	var date time.Time

	if task.Date == "" {
		task.Date = time.Now().Format("20060102") // указзываем сегодняшнюю дату
	}
	date, err = time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, "некорректный формат даты", http.StatusBadRequest)
		return
	}

	now := time.Now()

	//  если дата меньше сегодняшней
	if date.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102") // Устанавливаем сегодняшнюю дату
		} else {
			// вычисляем следующую дату NextDate
			nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, "Ошибка вычисления следующей даты", http.StatusInternalServerError)
				return
			}
			task.Date = nextDate
		}
	}

	respId, err := db.AddTask(c.db, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response := map[string]string{"id": fmt.Sprintf("%d", respId)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (c *TaskController) NextDateHandler(w http.ResponseWriter, r *http.Request) {

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "недостаточно значений", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "недопустимый формат даты", http.StatusBadRequest)
		return
	}

	nextDate, err := nextdate.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, nextDate)
}
