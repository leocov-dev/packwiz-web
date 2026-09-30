<script setup lang="ts">

import {useAuthStore} from "@/stores/auth.ts";

const emit = defineEmits(['close'])

// accounts created through an identity provider have no password yet: they
// set a first one instead of changing an existing one
const {hasPassword = true} = defineProps<{ hasPassword?: boolean }>()

const form = ref()

const isValid = ref(false)

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const showOldPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

const error = ref('')

const rules = {
  required: (value: string) => !!value || 'This field is required.',
  checkAgainstOld: (value: string) => !hasPassword || value !== oldPassword.value || 'Passwords must be different.',
  checkAgainstNew: (value: string) => value === newPassword.value || 'Does not match new password.',
}

const authStore = useAuthStore()

const submitForm = async () => {
  if (hasPassword) {
    error.value = await authStore.changePassword(oldPassword.value, newPassword.value)
    return
  }

  error.value = await authStore.setPassword(newPassword.value)
  if (!error.value) emit('close')
}

watch([
  oldPassword,
  newPassword,
  confirmPassword,
], async () => {
  isValid.value = (await form.value.validate()).valid
})

</script>

<template>
  <v-card
    prepend-icon="mdi-lock"
    :title="hasPassword ? 'Change Password' : 'Set Password'"
    :subtitle="hasPassword
      ? 'You will be logged out after changing your password.'
      : 'Add a password so you can also sign in without your identity provider.'"
  >
    <v-divider class="mt-3" />
    <v-form
      ref="form"
      validate-on="input"
      @submit.prevent="submitForm"
    >
      <v-card-text class="px-4">
        <input
          id="username"
          type="text"
          autocomplete="username"
          hidden
        >
        <v-text-field
          v-if="hasPassword"
          v-model.trim="oldPassword"
          label="Old Password"
          :rules="[rules.required]"
          :append-inner-icon="showOldPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :type="showOldPassword ? 'text' : 'password'"
          autocomplete="current-password"
          @click:append-inner="showOldPassword = !showOldPassword"
        />
        <v-text-field
          v-model.trim="newPassword"
          label="New Password"
          :rules="[rules.required, rules.checkAgainstOld]"
          :append-inner-icon="showNewPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :type="showNewPassword ? 'text' : 'password'"
          autocomplete="new-password"
          @click:append-inner="showNewPassword = !showNewPassword"
        />
        <v-text-field
          v-model.trim="confirmPassword"
          label="Confirm Password"
          :rules="[rules.required, rules.checkAgainstNew]"
          :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :type="showConfirmPassword ? 'text' : 'password'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
        />

        <v-alert
          v-if="error"
          :text="error"
          variant="tonal"
          color="error"
        />
      </v-card-text>

      <v-divider />

      <v-card-actions class="ma-2">
        <v-btn
          min-width="120"
          text="Cancel"
          variant="text"
          @click="emit('close')"
        />

        <v-spacer />

        <v-btn
          type="submit"
          min-width="120"
          :disabled="!isValid"
          color="primary"
          :variant="isValid ? 'flat' : 'text'"
          text="Save"
        />
      </v-card-actions>
    </v-form>
  </v-card>
</template>

