<script setup lang="ts">

const model = defineModel<boolean>({required: true})

defineProps({
  maxWidth: {
    type: Number,
    default: 500,
  },
  title: {
    type: String,
    default: "Confirm Operation",
  },
  text: {
    type: String,
    default: "Are you sure you want to proceed?",
  },
  cancelText: {
    type: String,
    default: "Cancel",
  },
  acceptText: {
    type: String,
    default: "Accept",
  },
})

const emit = defineEmits(["result", "accepted", "canceled"])

const onCancel = () => {
  emit("canceled")
  emit("result", false)
  model.value = false
}

const onAccept = () => {
  emit("accepted")
  emit("result", true)
  model.value = false
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
      <v-card-text class="text-pre-line">
        {{ text }}
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          class="me-3"
          color="primary"
          variant="tonal"
          @click="onCancel"
        >
          {{ cancelText }}
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          @click="onAccept"
        >
          {{ acceptText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
