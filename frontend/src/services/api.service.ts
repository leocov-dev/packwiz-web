import axios from "axios"
import router from "@/router";
import { useSnackbarStore } from "@/stores/snackbar";
import {useAuthStore} from "@/stores/auth";

export const baseUrl = import.meta.env.VITE_API_BASE_URL || window.location.origin;


export const apiClient = axios.create({
  baseURL: `${baseUrl}/api/`,
  timeout: 10000,
  withCredentials: true,
});

apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    if (error.response && error.response.status === 401) {

      if (error.config && error.config.url === 'v1/user/current') {
        return Promise.reject(error);
      }

      if (router.currentRoute.value.path !== '/auth/login') {
        const snackbarStore = useSnackbarStore();
        snackbarStore.showSnackbar('Authorization Error...', 'error');

        const authStore = useAuthStore();
        await authStore.logout(true);

      }

      return Promise.reject(error);
    }

    // Other errors
    return Promise.reject(error);
  }
);
