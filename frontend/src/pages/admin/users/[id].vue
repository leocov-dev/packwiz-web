<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute, useRouter} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {User} from "@/interfaces/user.ts";
import {deactivateUserById, fetchUserById, reactivateUserById} from "@/services/user.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {AxiosError} from "axios";

const route = useRoute<'/admin/users/[id]'>()
const router = useRouter()
const snackbarStore = useSnackbarStore()

const userId = computed(() => Number(route.params.id))

const {
  isLoading,
  data: user,
  error,
  reload,
} = buildDataLoader<User>(async () => {
  return fetchUserById(userId.value)
})

const showDeactivateDialog = ref(false)
const showReactivateDialog = ref(false)
const statusLoading = ref(false)

const handleError = (e: unknown) => {
  let msg = "Unknown error"
  if (e instanceof AxiosError) {
    msg = e.response?.data?.msg || "Unknown error"
  }
  snackbarStore.showSnackbar(msg, "error")
}

const deactivate = async () => {
  statusLoading.value = true
  try {
    await deactivateUserById(userId.value)
    await reload()
  } catch (e) {
    handleError(e)
  } finally {
    statusLoading.value = false
  }
}

const reactivate = async () => {
  statusLoading.value = true
  try {
    await reactivateUserById(userId.value)
    await reload()
  } catch (e) {
    handleError(e)
  } finally {
    statusLoading.value = false
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

    <v-card v-else>
      <ConfirmationDialog
        v-model="showDeactivateDialog"
        title="Confirm Deactivate User"
        text="Are you sure you want to deactivate this user?
        They will be logged out and unable to log in until reactivated."
        @accepted="deactivate"
      />
      <ConfirmationDialog
        v-model="showReactivateDialog"
        title="Confirm Reactivate User"
        text="Are you sure you want to reactivate this user?
        They will be able to log in again."
        @accepted="reactivate"
      />

      <v-card-title>
        <div class="d-flex align-center ga-3">
          <h1 class="me-5">
            {{ user.fullName }}
          </h1>
          <v-chip
            v-if="user.isAdmin"
            :text="user.username === 'admin' ? 'Default Admin Account' : 'Admin Account'"
            color="warning"
          />
          <v-chip
            :color="user.isActive ? 'success' : 'error'"
            :text="user.isActive ? 'Active' : 'Deactivated'"
          />
        </div>
      </v-card-title>

      <v-divider />

      <v-list>
        <v-list-item
          title="Username"
          :subtitle="user.username"
        />
        <v-list-item
          title="Email"
          :subtitle="user.email"
        />
        <v-list-item
          title="Created"
          :subtitle="new Date(user.createdAt).toLocaleString()"
        />
        <v-list-item
          title="Updated"
          :subtitle="new Date(user.updatedAt).toLocaleString()"
        />
      </v-list>

      <v-card-actions class="ma-3">
        <v-btn
          text="Edit"
          @click="router.push(`/admin/users/${userId}/edit`)"
        />
        <v-btn
          v-if="user.isActive"
          text="Deactivate"
          color="error"
          :loading="statusLoading"
          :disabled="statusLoading"
          @click="showDeactivateDialog = true"
        />
        <v-btn
          v-else
          text="Reactivate"
          color="success"
          :loading="statusLoading"
          :disabled="statusLoading"
          @click="showReactivateDialog = true"
        />
      </v-card-actions>
    </v-card>
  </div>
</template>
