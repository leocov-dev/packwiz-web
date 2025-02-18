import axios from "axios"

const isDevelopment = process.env.NODE_ENV === 'development'
const apiPort = process.env.PWW_PORT || "8080"
const baseUrl = isDevelopment ? `http://localhost:${apiPort}` : window.location.origin

export const apiClient = axios.create({
  baseURL: `${baseUrl}/api/`,
  timeout: 10000,
});
