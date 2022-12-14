package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	//Get priority sort parameter
	prior := ctx.Query("priority")

	// Get tasks in DB
	var tasks []database.Task
	query := "SELECT id, title, created_at, is_done, content, priority FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
	switch {
		case kw != "":
			if is_done != "" {
				err = db.Select(&tasks, /*"SELECT * FROM tasks WHERE title LIKE ? AND is_done = ?"*/
				query + " AND title LIKE ? AND is_done = ?", userID, "%" + kw + "%", is_done=="済")
			} else {
				err = db.Select(&tasks,query + " AND title LIKE ?" , userID, "%" + kw + "%")
			}
		case is_done!="":
			err = db.Select(&tasks, query + " AND is_done = ?", userID, is_done=="済")
		case prior!="":
			err = db.Select(&tasks, query + " ORDER BY priority DESC", userID)
		default:
			err = db.Select(&tasks, query, userID)
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
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
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
	query := "SELECT id, title, created_at, is_done, content, priority FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND id = ?"
	err = db.Get(&task, query, userID, id) // Use DB#Get for one entry
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
	//get user id
	userID := sessions.Default(ctx).Get("user")
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
	prior, ex3 := ctx.GetPostForm("priority")
	if !ex3 {
		Error(http.StatusBadRequest, "No priority is given")(ctx)
		return
	}
	priority, ex4 := strconv.Atoi(prior)
	if ex4 != nil {
		Error(http.StatusBadRequest, ex4.Error())(ctx)
		return
	} 
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	tx := db.MustBegin()
	//Create new data with given title on DB
	result, err := tx.Exec("INSERT INTO tasks (title, content, priority) VALUES (?, ?, ?)", title, content, priority)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//get task id
	taskID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	//preserve user and task id  to ownership
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()

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
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
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
	query := "SELECT id, title, created_at, is_done, content, priority FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND id = ?"
	err = db.Get(&task, query, userID, id)
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
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
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
	prior, ex4 := ctx.GetPostForm("priority")
	if !ex4 {
		Error(http.StatusBadRequest, "No priority is given")(ctx)
		return
	}
	priority, ex5 := strconv.Atoi(prior)
	if ex5 != nil {
		Error(http.StatusBadRequest, ex5.Error())(ctx)
		return
	} 
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx := db.MustBegin()
	//update data with given title on DB
	//check task is present
    var cnt int
    err = tx.Get(&cnt, "SELECT COUNT(*) FROM ownership WHERE user_id=? AND task_id=?", userID, id)
    if err != nil {
		tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if cnt == 0 {
		tx.Rollback()
        Error(http.StatusBadRequest, "No task")(ctx)
        return
    }
	_, err = tx.Exec("UPDATE tasks SET title = ?, content = ?, is_done = ?, priority = ? WHERE id = ?", title, content, is_done, priority, id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

//delete the selected task
func DeleteTask(ctx *gin.Context) {
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
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
	tx := db.MustBegin()
	query := "DELETE FROM tasks WHERE id IN  (SELECT task_id FROM ownership WHERE user_id = ? AND task_id = ?)"
	_, err = tx.Exec(query, userID, id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("DELETE FROM ownership WHERE user_id = ? AND task_id = ?", userID, id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	//Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}

//show share task form
func CommonTaskForm(ctx *gin.Context) {
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
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
	query := "SELECT id, title, content, priority FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND id = ?"
	err = db.Get(&task, query, userID, id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Render edit form
	ctx.HTML(http.StatusOK, "share_task_form.html",
		gin.H{"Title": task.Title, "Content": task.Content, "ID": task.ID, "Priority": task.Priority})
}

//share task to other people
func ShareTask(ctx *gin.Context) {
	//get user id to distinguish user task
	userID := sessions.Default(ctx).Get("user")
	//get id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	user_name, ex1 := ctx.GetPostForm("user_name")
	if !ex1 {
		Error(http.StatusBadRequest, "No user_name is given")(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//check user_id is present
	tx := db.MustBegin()
	var user_id int
    err = tx.Get(&user_id, "SELECT id FROM users WHERE name=?", user_name)
	if err != nil {
		//Get target task
		var task database.Task
		query := "SELECT id, title, content, priority FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND id = ?"
		err = db.Get(&task, query, userID, id)
		if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}
		tx.Rollback()
        ctx.HTML(http.StatusBadRequest, "share_task_form.html", 
		        gin.H{"Error": "Username is invalid", "ID" : id, "Title" : task.Title, "Content" : task.Content, "Priority" : task.Priority})
        return
    }
	//register task to designated user
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", user_id, id)
	if err != nil {
		tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	//Render status
	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", id))
}
