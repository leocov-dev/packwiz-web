<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRouter} from "vue-router";
import {createUser} from "@/services/user.service.ts";
import AdminUserCreateForm, {type CreateUserFormData} from "@/components/user/AdminUserCreateForm.vue";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {AxiosError} from "axios";

const router = useRouter()
const snackbarStore = useSnackbarStore()

const onCreate = async (userData: CreateUserFormData) => {
  try {
    const user = await createUser(userData)
    await router.push(`/admin/users/${user.id}`)
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
    <AdminUserCreateForm @create-user="onCreate" />
  </div>
</template>
