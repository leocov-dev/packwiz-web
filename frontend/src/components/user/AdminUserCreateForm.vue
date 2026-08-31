<script lang="ts">
export interface CreateUserFormData {
  username: string,
  fullName: string,
  email: string,
  password: string,
  isAdmin: boolean,
}
</script>

<script setup lang="ts">
const event = defineEmits(['create-user'])

const isValid = ref(false)

const form = ref()

const showPassword = ref(false)

const formModel = reactive<CreateUserFormData>({
  username: '',
  fullName: '',
  email: '',
  password: '',
  isAdmin: false,
})

const rules = {
  usernameRequired: (value: string) => !!value || "Username is required",
  nameRequired: (value: string) => !!value || "Name is required",
  emailRequired: (value: string) => !!value || "Email is required",
  emailValid: (value: string) => {
    const pattern = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
    return pattern.test(value) || "Invalid email"
  },
  passwordRequired: (value: string) => !!value || "Password is required",
  passwordLength: (value: string) => (value.length >= 12 && value.length <= 64) || "Password must be 12-64 characters",
}
</script>

<template>
  <v-card>
    <v-card-title>
      <h1 class="me-5">
        New User
      </h1>
    </v-card-title>

    <v-divider />

    <v-form
      ref="form"
      v-model="isValid"
      class="ma-6"
      @submit.prevent="event('create-user', formModel)"
    >
      <v-text-field
        v-model.trim="formModel.username"
        label="Username"
        :rules="[rules.usernameRequired]"
        autocomplete="username"
      />
      <v-text-field
        v-model.trim="formModel.fullName"
        label="Name"
        :rules="[rules.nameRequired]"
      />
      <v-text-field
        v-model.trim="formModel.email"
        label="Email"
        :rules="[
          rules.emailRequired,
          rules.emailValid,
        ]"
      />
      <v-text-field
        v-model="formModel.password"
        label="Password"
        :rules="[
          rules.passwordRequired,
          rules.passwordLength,
        ]"
        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :type="showPassword ? 'text' : 'password'"
        autocomplete="new-password"
        @click:append-inner="showPassword = !showPassword"
      />
      <v-checkbox
        v-model="formModel.isAdmin"
        label="Grant admin access"
      />

      <div class="d-flex justify-end ga-4">
        <v-btn
          text="Create"
          type="submit"
          min-width="120"
          :disabled="!isValid"
        />
      </div>
    </v-form>
  </v-card>
</template>
