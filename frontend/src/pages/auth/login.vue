<route lang="yaml">
meta:
  noAuth: true
  layout: auth
</route>

<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";
import type {OidcPublicProvider} from "@/interfaces/oidc.ts";
import {fetchAuthConfig, oidcLoginUrl} from "@/services/oidc.service.ts";
import {oidcErrorMessage} from "@/lib/oidc-errors.ts";

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const showPassword = ref(false);
const username = ref('');
const password = ref('');
const loading = ref(false);
const errorMessage = ref(oidcErrorMessage(route.query.error) ?? '');
const providers = ref<OidcPublicProvider[]>([]);

const rules = {
  required: (value: string) => !!value || 'This field is required',
};

const submitForm = async () => {
  errorMessage.value = '';

  if (!username.value || !password.value) return;

  loading.value = true;

  try {
    await authStore.login(username.value, password.value);
    await redirect();
  } catch {
    errorMessage.value = 'Username or password is incorrect.';
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

const loginWithProvider = (provider: OidcPublicProvider) => {
  const target = route.query.redirect
  const redirectPath = typeof target === 'string' && !target.endsWith('logout') ? target : undefined
  // full-page navigation: the browser leaves for the identity provider
  window.location.assign(oidcLoginUrl(provider.slug, redirectPath))
}

onMounted(async () => {
  if (authStore.isAuthenticated) {
    return redirect();
  }

  try {
    providers.value = await fetchAuthConfig()
  } catch {
    // local login must keep working if the provider list cannot be loaded
    providers.value = []
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
        height="130"
        src="/packwiz.png"
      />

      <div class="mb-5 text-center">
        <h2 class="font-weight-bold">
          Packwiz Web
        </h2>
      </div>

      <v-card
        v-if="errorMessage"
        class="mb-5"
        color="error"
        variant="tonal"
      >
        <v-card-text>
          {{ errorMessage }}
        </v-card-text>
      </v-card>

      <v-form
        @submit.prevent="submitForm"
      >
        <v-text-field
          v-model.trim="username"
          label="Username"
          :rules="[rules.required]"
          autofocus
          autocomplete="username"
        />

        <v-text-field
          v-model.trim="password"
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

      <template v-if="providers.length">
        <div class="d-flex align-center ga-3 my-4 text-medium-emphasis">
          <v-divider />
          <span class="text-caption">or</span>
          <v-divider />
        </div>

        <v-btn
          v-for="provider in providers"
          :key="provider.slug"
          class="mb-2"
          variant="tonal"
          prepend-icon="mdi-login"
          :text="`Sign in with ${provider.displayName}`"
          block
          @click="loginWithProvider(provider)"
        />
      </template>
    </v-responsive>
  </v-container>
</template>
