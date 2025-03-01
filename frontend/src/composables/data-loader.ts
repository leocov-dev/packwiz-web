import {onMounted, ref} from "vue";

export function buildDataLoader<R>(func: () => Promise<R>) {
  const isLoading = ref(true);
  const data = ref<R>();
  const error = ref<any | null>(null);

  const reload = async () => {
    isLoading.value = true;
    error.value = null;
    try {
      data.value = await func()
    } catch (e) {
      error.value = e;
    }
    isLoading.value = false;
  };

  onMounted(async () => {
    await reload();
  });

  return {
    isLoading,
    data,
    reload,
    error,
  };
}
