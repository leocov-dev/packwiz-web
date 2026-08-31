<script setup lang="ts">
import {PackPermission, type PackCollaborator} from "@/interfaces/pack.ts";
import {removeCollaborator, updateCollaboratorPermission} from "@/services/packs.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import axios from "axios";

const {packId, collaborator} = defineProps<{ packId: number, collaborator: PackCollaborator }>()

const emit = defineEmits(['changed'])

const showRemoveDialog = ref(false)
const loading = ref(false)
const error = ref(false)
const errorMsg = ref("")

const permissionItems = [
  {title: 'Static (link only)', value: PackPermission.STATIC},
  {title: 'View', value: PackPermission.VIEW},
  {title: 'Edit', value: PackPermission.EDIT},
]

const handleError = (e: unknown, fallback: string) => {
  error.value = true
  if (axios.isAxiosError(e)) {
    errorMsg.value = e.response?.data?.error || fallback
  } else {
    errorMsg.value = String(e)
  }
}

const onPermissionChange = async (value: PackPermission) => {
  loading.value = true
  error.value = false
  try {
    await updateCollaboratorPermission(packId, collaborator.userId, value)
    emit('changed')
  } catch (e) {
    handleError(e, "Failed to update permission")
  } finally {
    loading.value = false
  }
}

const onRemove = async () => {
  loading.value = true
  error.value = false
  try {
    await removeCollaborator(packId, collaborator.userId)
    emit('changed')
  } catch (e) {
    handleError(e, "Failed to remove collaborator")
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <v-card class="ma-1 ps-5 pe-5 pt-3 pb-3 elevation-4">
    <ConfirmationDialog
      v-model="showRemoveDialog"
      title="Remove Collaborator"
      :text="`Are you sure you want to remove ${collaborator.username} from this pack?`"
      accept-text="Remove"
      @accepted="onRemove"
    />

    <v-alert
      v-if="error"
      class="mb-3"
      :text="'Error: ' + (errorMsg || 'something went wrong')"
      type="error"
      icon="mdi-alert"
      density="compact"
      closable
      @click:close="error = false"
    />

    <div class="d-flex align-center flex-wrap">
      <div>
        <div>{{ collaborator.fullName || collaborator.username }}</div>
        <div class="text-subtitle-2 text-disabled">
          {{ collaborator.email }}
        </div>
      </div>

      <v-spacer />

      <v-select
        v-tooltip="'Static: install/update via personal link only, no dashboard access'"
        :model-value="collaborator.permission"
        :items="permissionItems"
        label="Permission"
        density="compact"
        variant="outlined"
        hide-details
        max-width="220"
        class="me-3"
        :disabled="loading"
        @update:model-value="onPermissionChange"
      />

      <v-btn
        icon="mdi-account-remove"
        density="comfortable"
        color="error"
        variant="outlined"
        :disabled="loading"
        @click="showRemoveDialog = true"
      />
    </div>
  </v-card>
</template>
