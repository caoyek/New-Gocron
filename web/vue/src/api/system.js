import httpClient from '../utils/httpClient'

export default {
  loginLogList (query, callback) {
    httpClient.get('/system/login-log', query, callback)
  },
  loginSecurity (callback) {
    httpClient.get('/system/login-security', {}, callback)
  },
  updateLoginSecurity (data, callback) {
    httpClient.post('/system/login-security/update', data, callback)
  },
  removeLoginBlock (id, callback) {
    httpClient.post(`/system/login-security/block/remove/${id}`, {}, callback)
  }
}
