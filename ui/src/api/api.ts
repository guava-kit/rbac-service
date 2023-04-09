import { Axios } from 'axios'

const BASE_URL = '/api'

export function createAxios() {
  const axios = new Axios({ baseURL: BASE_URL })
  axios.interceptors.request.use((config) => {
    return config
  })
  axios.interceptors.response.use(
    (response) => {
      return response
    },
    (error) => {
      return Promise.reject(error)
    }
  )
  return axios
}

export const axios = createAxios()
