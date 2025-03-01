<route lang="yaml">
meta:
  noAuth: true
  layout: auth
</route>

<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const showPassword = ref(false);
const username = ref('');
const password = ref('');
const loading = ref(false);
const showError = ref(false);

const rules = {
  required: (value: string) => !!value || 'This field is required',
};

const submitForm = async () => {
  showError.value = false;

  if (!username.value || !password.value) return;

  loading.value = true;

  try {
    await authStore.login(username.value, password.value);
    await redirect();
  } catch {
    showError.value = true;
  } finally {
    loading.value = false;
  }

}

const redirect = async () => {
  if ((route.query.redirect as string)?.endsWith("logout")){
    return router.push("/");
  }
  return router.push((route.query.redirect as string) || '/');
}

onMounted(async () => {
  if (authStore.isAuthenticated) {
    return redirect();
  }
})

</script>

<template>
  <v-container class="fill-height mt-n3">
    <v-responsive
      class="align-center fill-height mr-auto ml-auto"
      max-width="300"
    >
      <v-img
        class="mb-4"
        height="100"
        src="@/assets/logo.png"
      />

      <div class="mb-5 text-center">
        <h2 class="font-weight-bold">
          Packwiz Web
        </h2>
      </div>

      <v-card
        v-if="showError"
        class="mb-5"
        color="error"
        variant="tonal"
      >
        <v-card-text>
          Username or password is incorrect.
        </v-card-text>
      </v-card>

      <v-form
        @submit.prevent="submitForm"
      >
        <v-text-field
          v-model="username"
          label="Username"
          :rules="[rules.required]"
          autofocus
          autocomplete="username"
        />

        <v-text-field
          v-model="password"
          :type="showPassword ? 'text' : 'password'"
          label="Password"
          :rules="[rules.required]"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          autocomplete="current-password"
          @click:append-inner="showPassword = !showPassword"
        />

        <v-btn
          class="mt-2"
          :disabled="loading"
          :loading="loading"
          color="primary"
          type="submit"
          block
        >
          Login
        </v-btn>
      </v-form>
    </v-responsive>
  </v-container>
</template>
