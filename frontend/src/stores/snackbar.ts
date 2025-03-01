import { defineStore } from 'pinia';

export const useSnackbarStore = defineStore('snackbar', {
  state: () => ({
    show: false,
    message: '',
    color: 'info', // Default to 'info' (use 'error', 'success', etc.)
    timeout: 5000, // 5 seconds
  }),
  actions: {
    showSnackbar(message: string, color = 'info', timeout = 5000): void {
      this.show = true;
      this.message = message;
      this.color = color;
      this.timeout = timeout;
    },
    closeSnackbar() {
      this.show = false;
      this.message = '';
      this.color = 'info';
    },
  },
});
