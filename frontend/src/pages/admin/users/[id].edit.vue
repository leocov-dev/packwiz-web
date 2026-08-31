<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute, useRouter} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {User} from "@/interfaces/user.ts";
import {fetchUserById, updateUserById} from "@/services/user.service.ts";
import AdminUserEditForm, {type UserProfileFormData} from "@/components/user/AdminUserEditForm.vue";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {AxiosError} from "axios";

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

    <AdminUserEditForm
      v-else
      :user="user"
      @update-user="onUpdate"
    />
  </div>
</template>
