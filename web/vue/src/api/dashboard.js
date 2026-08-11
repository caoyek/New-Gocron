import httpClient from '../utils/httpClient'

export default {
  index (callback) {
    httpClient.get('/dashboard', {}, callback)
  }
}
