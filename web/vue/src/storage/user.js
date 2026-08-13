class User {
  get () {
    return {
      'token': this.getToken(),
      'uid': this.getUid(),
      'username': this.getUsername(),
      'isAdmin': this.getIsAdmin()
    }
  }

  getToken () {
    return localStorage.getItem('token') || ''
  }

  setToken (token) {
    localStorage.setItem('token', token)
    return this
  }

  clear () {
    const pinnedTasks = {}
    for (let index = 0; index < localStorage.length; index++) {
      const key = localStorage.key(index)
      if (key && key.indexOf('new-gocron:pinned-tasks:') === 0) {
        pinnedTasks[key] = localStorage.getItem(key)
      }
    }
    localStorage.clear()
    Object.keys(pinnedTasks).forEach(key => localStorage.setItem(key, pinnedTasks[key]))
  }

  getUid () {
    return localStorage.getItem('uid') || ''
  }

  setUid (uid) {
    localStorage.setItem('uid', uid)
    return this
  }

  getUsername () {
    return localStorage.getItem('username') || ''
  }

  setUsername (username) {
    localStorage.setItem('username', username)
    return this
  }

  getIsAdmin () {
    let isAdmin = localStorage.getItem('is_admin')
    return isAdmin === '1'
  }

  setIsAdmin (isAdmin) {
    localStorage.setItem('is_admin', isAdmin)
    return this
  }
}

export default new User()
