import axios from "axios"

const isDevelopment = process.env.NODE_ENV === 'development'
const baseUrl = isDevelopment ? 'http://localhost:8080' : window.location.origin

export const apiClient = axios.create({
  baseURL: `${baseUrl}/api/`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json'
  },
});
