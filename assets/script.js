/* placeholder file for JavaScript */
//delete alert
const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}

//task update alert
const confirm_update = (id) => {
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) return true;
    else return false;
}

//logout alert
const confirm_logout = () => {
    if(window.confirm(`ログアウトします．よろしいですか？`)) {
        location.href = `/logout`;
    }
}

//withdraw alert
const confirm_withdraw = () => {
    if(window.confirm(`このアプリを退会します. よろしいですか？`)) {
        location.href = `/withdraw`;
    }
}

//update user information alert 
const confirm_user_update = () => {
    if(window.confirm(`ユーザー情報を更新します．よろしいですか？`)) return true;
    else return false;
}

//update share information
const confirm_share = (id) => {
    if(window.confirm(`Task ${id} を指定したユーザーへ共有します．よろしいですか？`)) return true;
    else return false;
}