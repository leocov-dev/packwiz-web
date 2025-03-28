<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";
import UserProfile, {type UserProfileFormData} from "@/components/user/UserProfile.vue";
import {updateCurrentUser} from "@/services/user.service.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {AxiosError} from "axios";

const authStore = useAuthStore()
const snackbarStore = useSnackbarStore()

const user = authStore.user

const onUpdate = async (userData: UserProfileFormData) => {
  if (!user) return

  try {
    await updateCurrentUser(userData)
  } catch (e) {
    let msg = "Unknown error"
    if (e instanceof AxiosError) {
      msg = e.response?.data?.msg || "Unknown error"
    }
    snackbarStore.showSnackbar(
      msg,
      "error"
    )
  }
}
</script>

<template>
  <div class="ma-6">
    <UserProfile
      v-if="user"
      :user="user"
      @update-user="onUpdate"
    />
  </div>
</template>

