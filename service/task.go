package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
    kw := ctx.Query("kw")
	// Get is_done or not is_done
	is_done := ctx.Query("is_done")

	// Get tasks in DB
	var tasks []database.Task
	switch {
		case kw != "":
			if is_done != "" {
				err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ? AND is_done = ?", "%" + kw + "%", is_done=="済")
			} else {
				err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%" + kw + "%")
			}
		case is_done!="":
			err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done = ?", is_done=="済")
		default:
			err = db.Select(&tasks, "SELECT * FROM tasks")
    }
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}

//show register new task form
func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_task_form.html", gin.H{"Title": "Task registration"})
}

//register task
func RegisterTask(ctx *gin.Context) {
	//Get task title and content
	title, ex1 := ctx.GetPostForm("title")
	if !ex1 {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	content, ex2 := ctx.GetPostForm("content")
	if !ex2 {
		Error(http.StatusBadRequest, "No content is given")(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Create new data with given title on DB
	result, err := db.Exec("INSERT INTO tasks (title, content) VALUES (?, ?)", title, content)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Render status
	path := "/list" // task list page for default
	if id, err := result.LastInsertId(); err == nil {
		//task id page when the result is correct
		path = fmt.Sprintf("/task/%d", id)
	}
	ctx.Redirect(http.StatusFound, path)
}

//show edit task form
func EditTaskForm(ctx *gin.Context) {
	//get id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Render edit form
	ctx.HTML(http.StatusOK, "edit_task_form.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

//update task
func UpdateTask(ctx *gin.Context) {
	//Get task title, is_done, content and id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	title, ex1 := ctx.GetPostForm("title")
	if !ex1 {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	content, ex2 := ctx.GetPostForm("content")
	if !ex2 {
		Error(http.StatusBadRequest, "No content is given")(ctx)
		return
	}
	done, ex3 := ctx.GetPostForm("is_done")
	if !ex3 {
		Error(http.StatusBadRequest, "No is_done is given")(ctx)
		return
	}
	is_done, err := strconv.ParseBool(done)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//update data with given title on DB
	_, err = db.Exec("UPDATE tasks SET title = ?, content = ?, is_done = ? WHERE id = ?", title, content, is_done, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

//delete the selected task
func DeleteTask(ctx *gin.Context) {
	//get ID
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Delete the task from DB
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}
