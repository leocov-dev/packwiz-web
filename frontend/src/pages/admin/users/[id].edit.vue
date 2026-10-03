<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute, useRouter} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {User} from "@/interfaces/user.ts";
import {fetchUserById, resetUserPassword, updateUserById} from "@/services/user.service.ts";
import AdminUserEditForm, {type UserProfileFormData} from "@/components/user/AdminUserEditForm.vue";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {AxiosError} from "axios";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import GeneratedPasswordDialog from "@/components/user/GeneratedPasswordDialog.vue";

const route = useRoute<'/admin/users/[id].edit'>()
const router = useRouter()
const snackbarStore = useSnackbarStore()

const userId = computed(() => Number(route.params.id))

const {
  isLoading,
  data: user,
  error,
} = buildDataLoader<User>(async () => {
  return fetchUserById(userId.value)
})

const confirmReset = ref(false)
const showPassword = ref(false)
const generatedPassword = ref("")

const onResetPassword = async () => {
  try {
    generatedPassword.value = await resetUserPassword(userId.value)
    showPassword.value = true
  } catch (e) {
    let msg = "Unknown error"
    if (e instanceof AxiosError) {
      msg = e.response?.data?.msg || "Unknown error"
    }
    snackbarStore.showSnackbar(msg, "error")
  }
}

const onPasswordDone = () => {
  generatedPassword.value = ""
}

const onUpdate = async (userData: UserProfileFormData) => {
  try {
    await updateUserById(userId.value, userData)
    await router.push(`/admin/users/${userId.value}`)
  } catch (e) {
    let msg = "Unknown error"
    if (e instanceof AxiosError) {
      msg = e.response?.data?.msg || "Unknown error"
    }
    snackbarStore.showSnackbar(msg, "error")
  }
}
</script>

<template>
  <div class="ma-6">
    <div
      v-if="isLoading"
    >
      <v-skeleton-loader
        elevation="0"
        theme="article"
        type="heading, subtitle, paragraph@2"
      />
    </div>

    <v-alert
      v-else-if="error || !user"
      type="error"
      icon="mdi-alert"
      text="Failed to load user."
    />

    <v-alert
      v-else-if="user.isSuperuser"
      type="info"
      variant="tonal"
      text="The default admin account cannot be edited."
    />

    <AdminUserEditForm
      v-else
      :user="user"
      @update-user="onUpdate"
    />

    <div
      v-if="user && !user.isSuperuser"
      class="mt-6"
    >
      <v-btn
        text="Reset password"
        prepend-icon="mdi-lock-reset"
        color="warning"
        variant="tonal"
        @click="confirmReset = true"
      />
    </div>

    <ConfirmationDialog
      v-model="confirmReset"
      title="Reset password"
      text="Generate a new random password for this user? Their current password stops working and they are signed out."
      accept-text="Reset"
      @accepted="onResetPassword"
    />

    <GeneratedPasswordDialog
      v-model="showPassword"
      :password="generatedPassword"
      title="Password reset"
      @done="onPasswordDone"
    />
  </div>
</template>
