<script setup lang="ts">
import type {UserIdentity} from "@/interfaces/oidc.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {fetchUserIdentities, unlinkUserIdentity} from "@/services/oidc.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";

const {userId} = defineProps<{ userId: number }>()

const snackbarStore = useSnackbarStore()

const {
  isLoading,
  data,
  reload,
} = buildDataLoader(() => fetchUserIdentities(userId))

const identities = computed(() => data.value ?? [])

const showUnlink = ref(false)
const target = ref<UserIdentity | null>(null)

const askUnlink = (identity: UserIdentity) => {
  target.value = identity
  showUnlink.value = true
}

const confirmUnlink = async () => {
  if (!target.value) return
  try {
    await unlinkUserIdentity(userId, target.value.id)
    snackbarStore.showSnackbar('Linked account removed.', 'success')
    await reload()
  } catch (e) {
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to unlink account.'), 'error')
  }
}
</script>

<template>
  <v-divider />

  <v-card-subtitle class="mt-4">
    Linked accounts
  </v-card-subtitle>

  <v-skeleton-loader
    v-if="isLoading"
    type="list-item-two-line"
  />

  <v-list v-else>
    <v-list-item
      v-for="identity in identities"
      :key="identity.id"
      :title="identity.providerName"
      :subtitle="identity.email || 'No email shared'"
    >
      <template #append>
        <v-btn
          text="Unlink"
          color="error"
          variant="text"
          density="comfortable"
          @click="askUnlink(identity)"
        />
      </template>
    </v-list-item>

    <v-list-item
      v-if="!identities.length"
      title="No linked accounts"
    />
  </v-list>

  <ConfirmationDialog
    v-model="showUnlink"
    title="Unlink Account"
    :text="`Remove the ${target?.providerName ?? ''} account from this user? It is refused if it is their only way to sign in.`"
    accept-text="Unlink"
    @accepted="confirmUnlink"
  />
</template>
