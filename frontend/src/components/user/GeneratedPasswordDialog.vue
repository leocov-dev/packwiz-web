<script setup lang="ts">
import {writeToClipboard} from "@/lib/clipboard.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const model = defineModel<boolean>({required: true})

defineProps({
  password: {
    type: String,
    required: true,
  },
  title: {
    type: String,
    default: "Password generated",
  },
})

const emit = defineEmits(["done"])

const snackbarStore = useSnackbarStore()

const copied = ref(false)
const showPassword = ref(false)

const onCopy = async (password: string) => {
  try {
    await writeToClipboard(password)
    copied.value = true
    snackbarStore.showSnackbar("Password copied to clipboard", "default", 2000)
  } catch {
    snackbarStore.showSnackbar("Unable to copy password", "error")
  }
}

const onDone = () => {
  model.value = false
  copied.value = false
  showPassword.value = false
  emit("done")
}
</script>

<template>
  <v-dialog
    v-model="model"
    persistent
    max-width="500"
  >
    <v-card class="pa-3">
      <v-card-title>
        {{ title }}
      </v-card-title>
      <v-card-text>
        <v-alert
          type="warning"
          variant="tonal"
          class="mb-4"
          text="This password is shown only once. Copy it now and give it to the user; they can change it after logging in."
        />
        <v-text-field
          :model-value="password"
          label="Password"
          readonly
          :type="showPassword ? 'text' : 'password'"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append-inner="showPassword = !showPassword"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          prepend-icon="mdi-content-copy"
          text="Copy"
          variant="tonal"
          @click="onCopy(password)"
        />
        <v-btn
          class="me-3"
          color="primary"
          text="Done"
          @click="onDone"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
