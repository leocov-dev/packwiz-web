<script setup lang="ts">
import type {OidcProvider} from "@/interfaces/oidc.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {deleteOidcProvider, fetchOrphanedUserCount, listOidcProviders} from "@/services/oidc.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {deleteConfirmationText, providerStatus} from "@/lib/oidc-form.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import OidcProviderDialog from "@/components/oidc/OidcProviderDialog.vue";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";

const snackbarStore = useSnackbarStore()

const headers = [
  {title: 'Name', key: 'displayName'},
  {title: 'Slug', key: 'slug'},
  {title: 'Issuer', key: 'issuerUrl'},
  {title: 'Status', key: 'enabled'},
  {title: 'New users', key: 'autoCreateUsers'},
  {title: '', key: 'actions', sortable: false, align: 'end' as const},
]

const {
  isLoading,
  data,
  reload,
} = buildDataLoader(listOidcProviders)

const providers = computed(() => data.value ?? [])

const showDialog = ref(false)
const editing = ref<OidcProvider | null>(null)

const showDelete = ref(false)
const deleting = ref<OidcProvider | null>(null)
const deleteText = ref('')

const openCreate = () => {
  editing.value = null
  showDialog.value = true
}

const openEdit = (provider: OidcProvider) => {
  editing.value = provider
  showDialog.value = true
}

const onRowClick = (_: Event, {item}: { item: OidcProvider }) => openEdit(item)

const askDelete = async (provider: OidcProvider) => {
  try {
    const orphaned = await fetchOrphanedUserCount(provider.id)
    deleteText.value = deleteConfirmationText(provider.displayName, orphaned)
    deleting.value = provider
    showDelete.value = true
  } catch (e) {
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to check affected users.'), 'error')
  }
}

const confirmDelete = async () => {
  if (!deleting.value) return
  try {
    await deleteOidcProvider(deleting.value.id)
    snackbarStore.showSnackbar('Provider deleted.', 'success')
    await reload()
  } catch (e) {
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to delete provider.'), 'error')
  }
}

const statusColor = (provider: OidcProvider) => {
  switch (providerStatus(provider)) {
    case 'broken':
      return 'error'
    case 'enabled':
      return 'success'
    default:
      return undefined
  }
}

const statusText = (provider: OidcProvider) => {
  const status = providerStatus(provider)
  return status.charAt(0).toUpperCase() + status.slice(1)
}
</script>

<template>
  <v-data-table
    :headers="headers"
    :items="providers"
    :loading="isLoading"
    :row-props="{class: 'cursor-pointer'}"
    @click:row="onRowClick"
  >
    <template #top>
      <v-toolbar
        class="ps-5 pe-5 pt-2 pb-2 d-flex flex-wrap"
        elevation="4"
      >
        <v-toolbar-title>OIDC Providers</v-toolbar-title>
        <v-btn
          text="New Provider"
          prepend-icon="mdi-plus"
          class="me-3"
          @click="openCreate"
        />
        <v-btn
          icon="mdi-refresh"
          aria-label="Refresh"
          @click="reload()"
        />
      </v-toolbar>
    </template>

    <template #[`item.enabled`]="{ item }">
      <v-chip
        :color="statusColor(item)"
        :text="statusText(item)"
        :prepend-icon="item.broken ? 'mdi-alert' : undefined"
        size="small"
        label
        variant="tonal"
      />
    </template>

    <template #[`item.autoCreateUsers`]="{ item }">
      <v-chip
        :color="item.autoCreateUsers ? 'warning' : undefined"
        :text="item.autoCreateUsers ? 'Auto-create' : 'Existing only'"
        size="small"
        label
        variant="tonal"
      />
    </template>

    <template #[`item.actions`]="{ item }">
      <v-btn
        icon="mdi-pencil"
        aria-label="Edit provider"
        variant="text"
        density="comfortable"
        @click.stop="openEdit(item)"
      />
      <v-btn
        icon="mdi-delete"
        aria-label="Delete provider"
        variant="text"
        density="comfortable"
        color="error"
        @click.stop="askDelete(item)"
      />
    </template>

    <template #loading>
      <v-skeleton-loader type="table-row@3" />
    </template>

    <template #no-data>
      <div class="d-flex justify-center ma-10">
        No OIDC providers configured. Local login is always available.
      </div>
    </template>
  </v-data-table>

  <OidcProviderDialog
    v-model="showDialog"
    :provider="editing"
    @saved="reload()"
  />

  <ConfirmationDialog
    v-model="showDelete"
    title="Delete OIDC Provider"
    :text="deleteText"
    accept-text="Delete"
    @accepted="confirmDelete"
  />
</template>
