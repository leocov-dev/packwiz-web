<script setup lang="ts">
import {buildDataLoader} from "@/composables/data-loader.ts"
import type {PackResponse} from "@/interfaces/pack.ts"
import type {PackWebhookResponse, PackWebhookTestResponse} from "@/interfaces/webhook.ts"
import {getPackWebhook, savePackWebhook, testPackWebhook} from "@/services/webhook.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {useSnackbarStore} from "@/stores/snackbar.ts"
import {
  buildWebhookRequest, formFromResponse, lastAttemptSummary, validateWebhookForm, validateWebhookUrl,
  WEBHOOK_METHODS, type WebhookFormData,
} from "@/lib/webhook.ts"

const {pack} = defineProps<{ pack: PackResponse }>()

const router = useRouter()
const snackbar = useSnackbarStore()

const {isLoading, data: saved, error: loadError} = buildDataLoader<PackWebhookResponse>(
  () => getPackWebhook(pack.id),
)

const form = ref<WebhookFormData | null>(null)
const saving = ref(false)
const testing = ref(false)
const error = ref<string | null>(null)
const testResult = ref<PackWebhookTestResponse | null>(null)

watch(saved, (value) => {
  if (value) form.value = formFromResponse(value)
}, {immediate: true})

const problem = computed(() => (form.value ? validateWebhookForm(form.value) : null))
const status = computed(() => (saved.value ? lastAttemptSummary(saved.value) : ""))

const addHeader = () => form.value?.headers.push({name: "", value: "", stored: false, visible: false})
const removeHeader = (index: number) => form.value?.headers.splice(index, 1)

const save = async () => {
  if (!form.value || problem.value) return
  error.value = null
  saving.value = true
  try {
    saved.value = await savePackWebhook(pack.id, buildWebhookRequest(form.value))
    snackbar.showSnackbar("Webhook saved", "success")
  } catch (e) {
    error.value = apiErrorMessage(e, "Error saving webhook...")
    console.error(e)
  } finally {
    saving.value = false
  }
}

// tests the form as entered, saved or not, so values can be checked before saving
const test = async () => {
  if (!form.value || problem.value) return
  error.value = null
  testResult.value = null
  testing.value = true
  try {
    testResult.value = await testPackWebhook(pack.id, buildWebhookRequest(form.value))
  } catch (e) {
    error.value = apiErrorMessage(e, "Error testing webhook...")
    console.error(e)
  } finally {
    testing.value = false
  }
}

const testSummary = computed(() => {
  const r = testResult.value
  if (!r) return ""
  return r.ok ? `Test succeeded (HTTP ${r.statusCode}).` : `Test failed: ${r.error}`
})

const back = () => router.push({path: `/packs/${pack.id}`})
</script>

<template>
  <div class="ma-6">
    <v-alert
      v-if="error"
      class="mb-6"
      :text="error"
      type="error"
      icon="mdi-alert"
      closable
    />

    <v-skeleton-loader
      v-if="isLoading"
      elevation="0"
      type="heading, subtitle, actions, paragraph@2"
    />

    <v-alert
      v-else-if="loadError || !form"
      type="error"
      icon="mdi-alert"
      text="Failed to load webhook."
    />

    <v-card v-else>
      <v-card-title class="d-flex align-center">
        <h1 class="me-5">
          Webhook
        </h1>
        <v-spacer />
        <v-btn
          prepend-icon="mdi-arrow-left"
          variant="text"
          text="Back to pack"
          @click="back"
        />
      </v-card-title>

      <v-form
        class="ma-6"
        @submit.prevent="save"
      >
        <p class="mb-4">
          Called after <strong>{{ pack.name }}</strong> changes (changes within a few seconds
          are combined into one call). The body is JSON describing the pack.
        </p>

        <v-alert
          v-if="saved?.broken"
          class="mb-4"
          type="warning"
          variant="tonal"
          text="The stored headers can no longer be decrypted. Enter them again."
        />

        <v-switch
          v-model="form.enabled"
          label="Enabled"
          color="primary"
          hide-details
        />

        <v-row class="mt-2">
          <v-col
            cols="12"
            sm="3"
          >
            <v-select
              v-model="form.method"
              label="Method"
              :items="WEBHOOK_METHODS"
            />
          </v-col>
          <v-col
            cols="12"
            sm="9"
          >
            <v-text-field
              v-model="form.url"
              label="URL"
              placeholder="https://example.com/hook"
              :rules="form.enabled || form.url ? [validateWebhookUrl] : []"
            />
          </v-col>
        </v-row>

        <v-divider class="mb-4" />

        <h2 class="text-h6 mb-1">
          Headers
        </h2>
        <p class="text-medium-emphasis mb-4">
          Use headers for credentials. Values are stored encrypted and are never shown again;
          leave a saved value blank to keep it. "Send test" calls the URL with the values
          entered here, without saving them.
        </p>

        <v-row
          v-for="(row, index) in form.headers"
          :key="index"
          dense
        >
          <v-col
            cols="12"
            sm="4"
          >
            <v-text-field
              v-model="row.name"
              label="Name"
              :readonly="row.stored"
            />
          </v-col>
          <v-col>
            <!-- not type=password, so browsers do not offer to save it; masked by CSS instead -->
            <v-text-field
              v-model="row.value"
              label="Value"
              autocomplete="off"
              spellcheck="false"
              data-1p-ignore
              data-lpignore="true"
              data-bwignore
              data-form-type="other"
              :class="{'webhook-secret': !row.visible}"
              :placeholder="row.stored ? 'Saved (unchanged)' : ''"
              :append-inner-icon="row.visible ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="row.visible = !row.visible"
            />
          </v-col>
          <v-col cols="auto">
            <v-btn
              icon="mdi-delete"
              variant="text"
              aria-label="Remove header"
              @click="removeHeader(index)"
            />
          </v-col>
        </v-row>

        <v-btn
          prepend-icon="mdi-plus"
          variant="text"
          text="Add header"
          @click="addHeader"
        />

        <v-divider class="my-4" />

        <v-alert
          :type="saved?.lastError ? 'error' : 'info'"
          variant="tonal"
          density="compact"
          class="mb-4"
          :text="status"
        />

        <v-alert
          v-if="testResult"
          :type="testResult.ok ? 'success' : 'error'"
          variant="tonal"
          density="compact"
          class="mb-4"
          :text="testSummary"
        />

        <v-alert
          v-if="problem"
          class="mb-4"
          type="warning"
          variant="tonal"
          density="compact"
          :text="problem"
        />

        <div class="d-flex justify-end ga-2">
          <v-btn
            :loading="testing"
            :disabled="saving || !!problem || !form.url.trim()"
            text="Send test"
            @click="test"
          />
          <v-btn
            type="submit"
            color="primary"
            :loading="saving"
            :disabled="!!problem || testing"
            text="Save"
          />
        </div>
      </v-form>
    </v-card>
  </div>
</template>

<style scoped>
.webhook-secret :deep(input) {
  -webkit-text-security: disc;
}
</style>
