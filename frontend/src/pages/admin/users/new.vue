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
import GeneratedPasswordDialog from "@/components/user/GeneratedPasswordDialog.vue";

const router = useRouter()
const snackbarStore = useSnackbarStore()

const generatedPassword = ref("")
const showPassword = ref(false)

const onPasswordDone = async () => {
  generatedPassword.value = ""
  await router.push('/admin/users')
}

const onCreate = async (userData: CreateUserFormData) => {
  try {
    const {generatedPassword: password} = await createUser(userData)
    generatedPassword.value = password
    showPassword.value = true
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

    <GeneratedPasswordDialog
      v-model="showPassword"
      :password="generatedPassword"
      title="User created"
      @done="onPasswordDone"
    />
  </div>
</template>
