import axios from "axios"
import router from "@/router";
import {useSnackbarStore} from "@/stores/snackbar";
import {useAuthStore} from "@/stores/auth.ts";

export const baseUrl = import.meta.env.VITE_API_BASE_URL || window.location.origin;


export const apiClient = axios.create({
  baseURL: `${baseUrl}/api/`,
  timeout: 60000,
  withCredentials: true,
});


apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    // FORBIDDEN ---------------------------------------------------------------
    if (error.response?.status === 403) {
      await router.push({path: "/"})
    }

    // UNAUTHORIZED ------------------------------------------------------------
    if (error.response?.status === 401 && error.config?.url !== 'v1/user') {
      if (router.currentRoute.value.path !== '/auth/login') {
        const snackbarStore = useSnackbarStore();
        snackbarStore.showSnackbar('Authorization Error...', 'error', 1500);

        const authStore = useAuthStore();
        await authStore.logout(false);
      }
    }

    return Promise.reject(error);
  }
);
