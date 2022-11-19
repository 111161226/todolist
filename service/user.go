package service

import (
	"crypto/sha256"
    "encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", nil)
}

//function for passward hash
func hash(pw string) []byte {
    const salt = "todolist.go/u_dogo#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}

//register user
func RegisterUser(ctx *gin.Context) {
    //receive form data
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
    pass_con := ctx.PostForm("pass_con")
    switch {
        case username == "":
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", 
            "Error": "Usernane is not provided", "Username": username})
            return
        case password == "":
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", 
            "Error": "Password is not provided", "Password": password})
            return
        case  pass_con== "":
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", 
            "Error": "Confirmed Password is not provided", "Pass_con": pass_con})
            return    
    }

    //check password is correct and has enough security
    en_len := 8
    switch {
        case password != pass_con:
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Confirmed Password doesn't match Password", "Username": username, "Password": password, "Pass_con":pass_con})
            return
        case len(password) < en_len:
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password doesn't have enough length", "Username": username, "Password": password, "Pass_con":pass_con})
            return
        default:
            //check security
            upperA := rune('A')
            upperZ := rune('Z')
            lowerA := rune('a')
            lowerZ := rune('z')
            small, big, num := false, false, false
            for  _, str := range(password) {
                if '0' <= str && str <= '9' {
                    num = true
                } else if upperA <= str && str <= upperZ{
                    big = true
                } else if lowerA <= str && str <= lowerZ{
                    small = true
                }
            }
            if(!(num && big && small)) {
                ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password doesn't include all these(number, uppercase and lowercase letter)", "Username": username, "Password": password, "Pass_con":pass_con})
                return
            }        
    }
    
    
    //database connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    //check duplicate
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password, "Pass_con": pass_con})
        return
    }
 
    //preserve the data into database
    result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    //confirm state of preservation
    id, _ := result.LastInsertId()
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.Redirect(http.StatusFound, "/list")
}

//show login form
func UserLoginForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "user_login.html", nil)
}

const userkey = "user"

//deal with login
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    //get user
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "user_login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    //check password is correct
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "user_login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    //preserve user id to session
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

//check user is login
func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

//log out
func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}


//delete
func Withdraw(ctx *gin.Context) {
    //get user id distinguish from others
    userID := sessions.Default(ctx).Get(userkey)
    //connect database
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    //Delete the task and user from DB
    tx := db.MustBegin()
	query := "DELETE FROM tasks WHERE id IN  (SELECT task_id FROM ownership WHERE user_id = ?)"
	_, err = tx.Exec(query, userID)
	if err != nil {
        tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    _, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    _, err = tx.Exec("DELETE FROM ownership WHERE user_id = ?", userID)
    if err != nil {
        tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    tx.Commit()

    //clear session information and Redirect to /
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}

//show edit user form
func EditUserForm(ctx *gin.Context) {
	//get user id to distinguish user 
	userID := sessions.Default(ctx).Get("user")
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Get target task
	var user database.User
	query := "SELECT name FROM users WHERE id = ?"
	err = db.Get(&user, query, userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Render edit form
	ctx.HTML(http.StatusOK, "edit_user_form.html", gin.H{"Name": user.Name})
}

//update user information
func UpdateUser(ctx *gin.Context) {
    //get user id to distinguish user 
	userID := sessions.Default(ctx).Get("user")
    //Get user name and password
	user_name, ex1 := ctx.GetPostForm("user_name")
	if !ex1 {
		Error(http.StatusBadRequest, "No username is given")(ctx)
		return
	}
	password, ex2 := ctx.GetPostForm("password")
	if !ex2 {
		Error(http.StatusBadRequest, "No password is given")(ctx)
		return
	}

    //Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    //check user duplicate
    tx := db.MustBegin()
    var duplicate int
    err = tx.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=? and id != ?", user_name, userID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        tx.Rollback()
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Error": "Username is already taken", "Name": user_name, "Password": password})
        return
    }
    //check password is correct and has enough security
    en_len := 8
    switch {
        case len(password) < en_len:
            ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Error": "Password doesn't have enough length", "Name": user_name, "Password": password})
            return
        default:
            //check security
            upperA := rune('A')
            upperZ := rune('Z')
            lowerA := rune('a')
            lowerZ := rune('z')
            small, big, num := false, false, false
            for  _, str := range(password) {
                if '0' <= str && str <= '9' {
                    num = true
                } else if upperA <= str && str <= upperZ{
                    big = true
                } else if lowerA <= str && str <= lowerZ{
                    small = true
                }
            }
            if(!(num && big && small)) {
                ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Error": "Password doesn't include all these(number, uppercase and lowercase letter)", "Name": user_name, "Password": password})
                return
            }        
    }
	//update data with given title on DB
	_, err = tx.Exec("UPDATE users SET name = ?, password = ? WHERE id = ?", user_name, hash(password), userID)
	if err != nil {
        tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    tx.Commit()
	ctx.Redirect(http.StatusFound, "/list")
}