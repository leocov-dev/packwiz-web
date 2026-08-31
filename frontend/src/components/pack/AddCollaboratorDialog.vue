<script setup lang="ts">
import {PackPermission, type UserSearchResult} from "@/interfaces/pack.ts";
import {addPackCollaborator, searchUsersForPack} from "@/services/packs.service.ts";
import axios from "axios";

const {packId, existingUserIds} = defineProps<{ packId: number, existingUserIds: number[] }>()

const model = defineModel<boolean>({required: true})

const emit = defineEmits(['added'])

const error = ref(false)
const errorMsg = ref("")
const loading = ref(false)
const searchLoading = ref(false)
const searchText = ref<string | undefined>(undefined)
const results = ref<UserSearchResult[]>([])
const selectedUserId = ref<number | null>(null)
const permission = ref<PackPermission>(PackPermission.VIEW)

const permissionItems = [
  {title: 'Static (link only)', value: PackPermission.STATIC},
  {title: 'View', value: PackPermission.VIEW},
  {title: 'Edit', value: PackPermission.EDIT},
]

const userItems = computed(() => results.value
  .filter(u => !existingUserIds.includes(u.userId))
  .map(u => ({title: `${u.username} (${u.email})`, value: u.userId})))

let debounceTimer: ReturnType<typeof setTimeout> | undefined
let requestSeq = 0

watch(searchText, (query: string | undefined) => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }

  if (!query || query.length < 2) {
    results.value = []
    searchLoading.value = false
    return
  }

  searchLoading.value = true
  const seq = ++requestSeq

  debounceTimer = setTimeout(async () => {
    try {
      const response = await searchUsersForPack(packId, query)
      if (seq === requestSeq) {
        results.value = response.users || []
      }
    } finally {
      if (seq === requestSeq) {
        searchLoading.value = false
      }
    }
  }, 400)
})

const reset = () => {
  error.value = false
  errorMsg.value = ""
  searchText.value = undefined
  results.value = []
  selectedUserId.value = null
  permission.value = PackPermission.VIEW
}

const submit = async () => {
  if (!selectedUserId.value) {
    return
  }

  error.value = false
  loading.value = true

  try {
    await addPackCollaborator(packId, selectedUserId.value, permission.value)
    model.value = false
    reset()
    emit('added')
  } catch (e) {
    error.value = true

    if (axios.isAxiosError(e)) {
      errorMsg.value = e.response?.data?.error || "Failed to add collaborator"
    } else {
      errorMsg.value = String(e)
    }
  } finally {
    loading.value = false
  }
}

const cancel = () => {
  model.value = false
  reset()
}
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="500"
    persistent
  >
    <v-card class="pa-3">
      <v-card-title>Add Collaborator</v-card-title>

      <v-card-text>
        <v-alert
          v-if="error"
          class="mb-4"
          :text="'Error: ' + (errorMsg || 'failed to add collaborator...')"
          type="error"
          icon="mdi-alert"
          closable
        />

        <v-autocomplete
          v-model="selectedUserId"
          v-model:search="searchText"
          :items="userItems"
          :loading="searchLoading"
          label="Search users by username, name, or email"
          no-data-text="Type at least 2 characters to search"
          clearable
        />

        <v-select
          v-model="permission"
          :items="permissionItems"
          label="Permission"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          class="me-3"
          variant="tonal"
          text="Cancel"
          :disabled="loading"
          @click="cancel"
        />
        <v-btn
          color="primary"
          variant="flat"
          text="Add"
          :disabled="loading || !selectedUserId"
          @click="submit"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
