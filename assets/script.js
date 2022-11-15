/* placeholder file for JavaScript */
const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirm_update = (id) => {
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) return true;
    else return false;
}

const confirm_logout = () => {
    if(window.confirm(`ログアウトします．よろしいですか？`)) {
        location.href = `/logout`;
    }
}

const confirm_withdraw = () => {
    if(window.confirm(`このアプリを退会します. よろしいですか？`)) {
        location.href = `/withdraw`;
    }
}

const confirm_user_update = () => {
    if(window.confirm(`ユーザー情報を更新します．よろしいですか？`)) return true;
    else return false;
}

const confirm_share = (id) => {
    if(window.confirm(`Task ${id} を指定したユーザーへ共有します．よろしいですか？`)) return true;
    else return false;
}