<script setup lang="ts">
import type {OidcPublicProvider, UserIdentity} from "@/interfaces/oidc.ts";
import {
  fetchAuthConfig,
  fetchMyIdentities,
  startLinkIdentity,
  unlinkMyIdentity,
} from "@/services/oidc.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {linkableProviders} from "@/lib/linked-accounts.ts";
import {oidcErrorMessage} from "@/lib/oidc-errors.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";

const route = useRoute()
const router = useRouter()
const snackbarStore = useSnackbarStore()

const isLoading = ref(true)
const identities = ref<UserIdentity[]>([])
const providers = ref<OidcPublicProvider[]>([])
const linkingSlug = ref('')

const showUnlink = ref(false)
const target = ref<UserIdentity | null>(null)

const available = computed(() => linkableProviders(providers.value, identities.value))

const load = async () => {
  isLoading.value = true
  try {
    const [mine, enabled] = await Promise.all([fetchMyIdentities(), fetchAuthConfig()])
    identities.value = mine
    providers.value = enabled
  } catch (e) {
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to load linked accounts.'), 'error')
  } finally {
    isLoading.value = false
  }
}

// the backend callback sends the browser back here with the outcome
const announceLinkResult = async () => {
  const error = oidcErrorMessage(route.query.oidcError)
  if (error) {
    snackbarStore.showSnackbar(error, 'error')
  } else if (route.query.oidc === 'linked') {
    snackbarStore.showSnackbar('Account linked.', 'success')
  } else {
    return
  }
  await router.replace({query: {}})
}

const link = async (provider: OidcPublicProvider) => {
  linkingSlug.value = provider.slug
  try {
    const url = await startLinkIdentity(provider.slug)
    // full-page navigation: the browser leaves for the identity provider
    window.location.assign(url)
  } catch (e) {
    linkingSlug.value = ''
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to start linking.'), 'error')
  }
}

const askUnlink = (identity: UserIdentity) => {
  target.value = identity
  showUnlink.value = true
}

const confirmUnlink = async () => {
  if (!target.value) return
  try {
    await unlinkMyIdentity(target.value.id)
    snackbarStore.showSnackbar('Account unlinked.', 'success')
    await load()
  } catch (e) {
    // includes the backend refusal to remove the last way to sign in
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to unlink account.'), 'error')
  }
}

onMounted(async () => {
  await announceLinkResult()
  await load()
})
</script>

<template>
  <v-card>
    <v-card-title>
      Linked Accounts
    </v-card-title>

    <v-divider />

    <v-skeleton-loader
      v-if="isLoading"
      type="list-item-two-line"
    />

    <template v-else>
      <v-list>
        <v-list-item
          v-for="identity in identities"
          :key="identity.id"
          :title="identity.providerName"
          :subtitle="`${identity.email || 'No email shared'} · linked ${new Date(identity.createdAt).toLocaleDateString()}`"
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

      <v-card-actions
        v-if="available.length"
        class="ma-3 flex-wrap ga-2"
      >
        <v-btn
          v-for="provider in available"
          :key="provider.slug"
          variant="tonal"
          prepend-icon="mdi-link-variant"
          :text="`Link ${provider.displayName}`"
          :loading="linkingSlug === provider.slug"
          :disabled="!!linkingSlug"
          @click="link(provider)"
        />
      </v-card-actions>
    </template>
  </v-card>

  <ConfirmationDialog
    v-model="showUnlink"
    title="Unlink Account"
    :text="`Unlink your ${target?.providerName ?? ''} account? You will no longer be able to sign in with it.`"
    accept-text="Unlink"
    @accepted="confirmUnlink"
  />
</template>
