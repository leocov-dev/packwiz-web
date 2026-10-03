import { defineStore } from 'pinia';

/** An optional action shown in the snackbar that navigates within the app. */
export interface SnackbarLink {
  text: string;
  to: string;
}

export const useSnackbarStore = defineStore('snackbar', {
  state: () => ({
    show: false,
    message: '',
    color: 'info', // Default to 'info' (use 'error', 'success', etc.)
    timeout: 5000, // 5 seconds
    link: null as SnackbarLink | null,
  }),
  actions: {
    showSnackbar(message: string, color = 'info', timeout = 5000, link: SnackbarLink | null = null): void {
      this.show = true;
      this.message = message;
      this.color = color;
      this.timeout = timeout;
      // always reset, so a link never carries over to the next message
      this.link = link;
    },
    closeSnackbar() {
      this.show = false;
      this.message = '';
      this.color = 'info';
      this.link = null;
    },
  },
});
