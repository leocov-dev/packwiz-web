<script setup lang="ts">
import type {PackCollaborator, PackRole} from "@/interfaces/pack.ts";
import {removeCollaborator, updateCollaboratorRole} from "@/services/packs.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import axios from "axios";
import {toTitleCase} from "@/services/utils.ts";

const {packId, collaborator, roles, canManage} = defineProps<{
  packId: number,
  collaborator: PackCollaborator,
  roles: PackRole[],
  canManage: boolean,
}>()

const emit = defineEmits(['changed'])

const showRemoveDialog = ref(false)
const loading = ref(false)
const error = ref(false)
const errorMsg = ref("")

// a role that is not in the assignable list (the pack owner) cannot be changed or removed
const isLocked = computed(() => !roles.some(r => r.id === collaborator.roleId))

const roleItems = computed(() => {
  const items = roles.map(r => ({title: toTitleCase(r.name), value: r.id}))
  if (isLocked.value) {
    items.push({title: toTitleCase(collaborator.roleName), value: collaborator.roleId})
  }
  return items
})

const handleError = (e: unknown, fallback: string) => {
  error.value = true
  if (axios.isAxiosError(e)) {
    errorMsg.value = e.response?.data?.error || fallback
  } else {
    errorMsg.value = String(e)
  }
}

const onRoleChange = async (value: number) => {
  loading.value = true
  error.value = false
  try {
    await updateCollaboratorRole(packId, collaborator.userId, value)
    emit('changed')
  } catch (e) {
    handleError(e, "Failed to update role")
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
        <div class="d-flex align-center ga-2">
          {{ collaborator.fullName || collaborator.username }}
          <v-chip
            v-if="collaborator.isActive === false"
            text="Deactivated"
            color="error"
            prepend-icon="mdi-account-off"
            size="small"
            label
            variant="tonal"
          />
        </div>
        <div class="text-subtitle-2 text-disabled">
          {{ collaborator.email }}
        </div>
      </div>

      <v-spacer />

      <v-select
        :model-value="collaborator.roleId"
        :items="roleItems"
        label="Role"
        density="compact"
        variant="outlined"
        hide-details
        max-width="220"
        class="me-3"
        :disabled="loading || isLocked || !canManage"
        @update:model-value="onRoleChange"
      />

      <v-btn
        v-if="canManage"
        icon="mdi-account-remove"
        density="comfortable"
        color="error"
        variant="text"
        :disabled="loading || isLocked"
        @click="showRemoveDialog = true"
      />
    </div>
  </v-card>
</template>
