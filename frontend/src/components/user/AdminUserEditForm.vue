<script lang="ts">
export interface UserProfileFormData {
  username: string,
  fullName: string,
  email: string,
}
</script>

<script setup lang="ts">
import type {User} from "@/interfaces/user.ts";

const {user} = defineProps<{ user: User }>()

const event = defineEmits(['update-user'])

const isValid = ref(true)

const form = ref()

const hasChanged = ref(false)

const formModel = reactive<UserProfileFormData>({
  username: user.username,
  fullName: user.fullName,
  email: user.email,
})

const rules = {
  usernameRequired: (value: string) => !!value || "Username is required",
  nameRequired: (value: string) => !!value || "Name is required",
  emailRequired: (value: string) => !!value || "Email is required",
  emailValid: (value: string) => {
    const pattern = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
    return pattern.test(value) || "Invalid email"
  },
}

const dataIsSame = () => {
  return user.email === formModel.email &&
    user.fullName === formModel.fullName &&
    user.username === formModel.username
}

const onCancel = () => {
  formModel.username = user.username
  formModel.fullName = user.fullName
  formModel.email = user.email
  setTimeout(
    () => {
      form.value.resetValidation()
    },
    100,
  )
}

watch(
  formModel,
  () => {
    hasChanged.value = !dataIsSame()
  },
)
</script>

<template>
  <v-card>
    <v-card-title>
      <h1 class="me-5">
        Edit User
      </h1>
    </v-card-title>

    <v-divider />

    <div
      v-if="user.isAdmin"
      class="d-flex justify-end ma-6"
    >
      <v-chip
        :text="user.username === 'admin' ? 'Default Admin Account' : 'Admin Account'"
        color="warning"
      />
    </div>

    <v-form
      ref="form"
      v-model="isValid"
      class="ma-6"
      @submit.prevent="event('update-user', formModel)"
    >
      <v-text-field
        v-model.trim="formModel.username"
        label="Username"
        :rules="[rules.usernameRequired]"
        :disabled="user.username === 'admin'"
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

      <div
        v-if="hasChanged"
        class="d-flex justify-end ga-4"
      >
        <v-btn
          text="Cancel"
          variant="tonal"
          min-width="120"
          @click="onCancel"
        />
        <v-btn
          text="Save"
          type="submit"
          min-width="120"
          :disabled="!isValid"
        />
      </div>
    </v-form>
  </v-card>
</template>
