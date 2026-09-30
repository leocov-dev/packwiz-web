<script setup lang="ts">
import type {OidcProvider} from "@/interfaces/oidc.ts";
import type {OidcProviderRequest} from "@/interfaces/requests.ts";
import {createOidcProvider, testOidcDiscovery, updateOidcProvider} from "@/services/oidc.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {
  DEFAULT_OIDC_SCOPES,
  normalizeScopes,
  slugify,
  validateIssuerUrl,
  validateScopes,
  validateSlug,
} from "@/lib/oidc-form.ts";

const model = defineModel<boolean>({required: true})

const {provider = null} = defineProps<{ provider?: OidcProvider | null }>()

const emit = defineEmits<{ saved: [provider: OidcProvider] }>()

const isEdit = computed(() => provider !== null)

const form = ref()
const isValid = ref<boolean | null>(null)
const saving = ref(false)
const testing = ref(false)
const error = ref('')
const discoveryMessage = ref('')
const discoveryOk = ref(false)
const slugTouched = ref(false)
const copied = ref(false)

const formModel = reactive<OidcProviderRequest>(emptyForm())

function emptyForm(): OidcProviderRequest {
  return {
    slug: '',
    displayName: '',
    issuerUrl: '',
    clientId: '',
    clientSecret: '',
    scopes: DEFAULT_OIDC_SCOPES,
    enabled: true,
    autoCreateUsers: false,
    linkByEmail: false,
  }
}

const resetForm = () => {
  Object.assign(formModel, emptyForm())
  if (provider) {
    Object.assign(formModel, {
      slug: provider.slug,
      displayName: provider.displayName,
      issuerUrl: provider.issuerUrl,
      clientId: provider.clientId,
      scopes: provider.scopes,
      enabled: provider.enabled,
      autoCreateUsers: provider.autoCreateUsers,
      linkByEmail: provider.linkByEmail,
    })
  }
  slugTouched.value = isEdit.value
  error.value = ''
  discoveryMessage.value = ''
  copied.value = false
}

watch(model, (open) => {
  if (open) resetForm()
}, {immediate: true})

watch(() => formModel.displayName, (name) => {
  if (!slugTouched.value) formModel.slug = slugify(name)
})

const rules = {
  required: (value: string) => !!value || 'This field is required.',
  slug: (value: string) => validateSlug(value),
  issuer: (value: string) => validateIssuerUrl(value),
  scopes: (value: string) => validateScopes(value),
}

const secretLabel = computed(() => isEdit.value && provider?.hasSecret
  ? 'Client secret (leave blank to keep the current one)'
  : 'Client secret')

const buildRequest = (): OidcProviderRequest => ({
  ...formModel,
  slug: formModel.slug.trim(),
  displayName: formModel.displayName.trim(),
  issuerUrl: formModel.issuerUrl.trim(),
  clientId: formModel.clientId.trim(),
  clientSecret: formModel.clientSecret.trim(),
  scopes: normalizeScopes(formModel.scopes),
})

const runDiscoveryTest = async () => {
  discoveryMessage.value = ''
  const check = validateIssuerUrl(formModel.issuerUrl)
  if (check !== true) {
    discoveryOk.value = false
    discoveryMessage.value = check
    return
  }

  testing.value = true
  try {
    const result = await testOidcDiscovery(formModel.issuerUrl.trim().replace(/\/+$/, ''))
    discoveryOk.value = true
    discoveryMessage.value = `Discovery succeeded. Authorization endpoint: ${result.authorizationEndpoint}`
  } catch (e) {
    discoveryOk.value = false
    discoveryMessage.value = apiErrorMessage(e, 'Discovery failed.')
  } finally {
    testing.value = false
  }
}

const submitForm = async () => {
  error.value = ''
  if (!(await form.value.validate()).valid) return

  saving.value = true
  try {
    const request = buildRequest()
    const saved = provider
      ? await updateOidcProvider(provider.id, request)
      : await createOidcProvider(request)
    emit('saved', saved)
    model.value = false
  } catch (e) {
    error.value = apiErrorMessage(e, 'Failed to save provider.')
  } finally {
    saving.value = false
  }
}

const copyRedirectUri = async () => {
  if (!provider?.redirectUri) return
  try {
    await navigator.clipboard.writeText(provider.redirectUri)
    copied.value = true
  } catch {
    copied.value = false
  }
}
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="640"
    persistent
    scrollable
  >
    <v-card
      prepend-icon="mdi-key-chain"
      :title="isEdit ? 'Edit OIDC Provider' : 'New OIDC Provider'"
    >
      <v-divider class="mt-3" />

      <v-card-text class="px-4">
        <v-form
          ref="form"
          v-model="isValid"
          validate-on="input"
          @submit.prevent="submitForm"
        >
          <v-alert
            v-if="provider?.broken"
            type="error"
            variant="tonal"
            class="mb-4"
            :text="provider.brokenError"
          />

          <v-text-field
            v-model.trim="formModel.displayName"
            label="Display name"
            hint="Shown on the login button."
            :rules="[rules.required]"
            autofocus
          />
          <v-text-field
            v-model.trim="formModel.slug"
            label="Slug"
            hint="Used in the callback URL. Changing it changes the redirect URI."
            persistent-hint
            :rules="[rules.slug]"
            @update:model-value="slugTouched = true"
          />
          <v-text-field
            v-model.trim="formModel.issuerUrl"
            class="mt-4"
            label="Issuer URL"
            hint="For example https://idp.example.com/realms/main or https://accounts.google.com"
            persistent-hint
            :rules="[rules.issuer]"
          >
            <template #append>
              <v-btn
                text="Test discovery"
                variant="tonal"
                :loading="testing"
                :disabled="testing"
                @click="runDiscoveryTest"
              />
            </template>
          </v-text-field>
          <v-alert
            v-if="discoveryMessage"
            :type="discoveryOk ? 'success' : 'error'"
            variant="tonal"
            class="mb-4"
            :text="discoveryMessage"
          />

          <v-text-field
            v-model.trim="formModel.clientId"
            class="mt-4"
            label="Client ID"
            :rules="[rules.required]"
            autocomplete="off"
          />
          <v-text-field
            v-model="formModel.clientSecret"
            :label="secretLabel"
            type="password"
            autocomplete="new-password"
          />
          <v-text-field
            v-model.trim="formModel.scopes"
            label="Scopes"
            hint="Space separated. Must include openid."
            persistent-hint
            :rules="[rules.scopes]"
          />

          <v-switch
            v-model="formModel.enabled"
            class="mt-4"
            label="Enabled"
            color="primary"
            hide-details
          />
          <v-switch
            v-model="formModel.linkByEmail"
            label="Link existing users by verified email"
            color="primary"
            hide-details
          />
          <v-switch
            v-model="formModel.autoCreateUsers"
            label="Create users automatically"
            color="primary"
            hide-details
          />
          <v-alert
            v-if="formModel.autoCreateUsers"
            type="warning"
            variant="tonal"
            class="mt-2"
            text="Anyone your IdP authenticates can get an account. Restrict access at the IdP."
          />

          <div class="mt-4">
            <v-text-field
              v-if="provider?.redirectUri"
              :model-value="provider.redirectUri"
              label="Redirect URI"
              hint="Register this exact URL with your identity provider."
              persistent-hint
              readonly
              append-inner-icon="mdi-content-copy"
              @click:append-inner="copyRedirectUri"
            />
            <v-alert
              v-else-if="isEdit"
              type="warning"
              variant="tonal"
              text="No redirect URI yet: set PWW_PUBLIC_URL on the server and restart. Sign-in is hidden until then."
            />
            <v-alert
              v-else
              type="info"
              variant="tonal"
              text="The redirect URI to register with your identity provider is shown after saving."
            />
            <div
              v-if="copied"
              class="text-caption text-success mt-1"
            >
              Copied to clipboard.
            </div>
          </div>

          <v-alert
            v-if="error"
            class="mt-4"
            type="error"
            variant="tonal"
            :text="error"
          />
        </v-form>
      </v-card-text>

      <v-divider />

      <v-card-actions class="ma-2">
        <v-btn
          min-width="120"
          text="Cancel"
          variant="text"
          @click="model = false"
        />

        <v-spacer />

        <v-btn
          min-width="120"
          color="primary"
          text="Save"
          :loading="saving"
          :disabled="isValid === false || saving"
          @click="submitForm"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
