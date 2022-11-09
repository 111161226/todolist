package service

import (
	"crypto/sha256"
	"net/http"

	"github.com/gin-gonic/gin"
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
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
            return
        case password == "":
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
            return
        case  pass_con== "":
            ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Confirmed Password is not provided", "Pass_con": pass_con})
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
                ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password doesn't include all these(number, uppercase and lowercase letter", "Username": username, "Password": password, "Pass_con":pass_con})
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
    err = db.Get(&user, "SELECT id, name, password, updated_at FROM users WHERE id = ?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.JSON(http.StatusOK, user)
}