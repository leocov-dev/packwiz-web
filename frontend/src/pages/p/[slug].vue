<route lang="yaml">
meta:
  noAuth: true
  layout: public
</route>

<script setup lang="ts">
import axios from "axios";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {downloadInstanceZip, fetchPublicPack, getClientSetupCommand} from "@/services/packs.service.ts";
import {writeToClipboard} from "@/lib/clipboard.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const route = useRoute()
const snackbar = useSnackbarStore()

const slug = computed(() => String((route.params as {slug: string}).slug))

const {data: pack, isLoading, error} = buildDataLoader(() => fetchPublicPack(slug.value))

const notFound = computed(() => axios.isAxiosError(error.value) && error.value.response?.status === 404)

const search = ref('')
const hideDependencies = ref(false)

const mods = computed(() => {
  const term = search.value.trim().toLowerCase()
  return (pack.value?.mods ?? []).filter(m =>
    (!hideDependencies.value || !m.isDependency)
    && (term === '' || m.name.toLowerCase().includes(term)),
  )
})

const headers = [
  {title: 'Name', key: 'name'},
  {title: 'Version', key: 'version'},
  {title: 'Type', key: 'type'},
  {title: 'Side', key: 'side'},
  {title: 'Source', key: 'source'},
]

const details = computed(() => {
  const p = pack.value
  if (!p) return []
  return [
    {label: 'Also accepts', value: p.acceptableGameVersions.join(', ')},
    {label: 'Pack format', value: p.packFormat},
    {label: 'Updated', value: new Date(p.updatedAt).toLocaleDateString()},
  ].filter(d => d.value)
})

const copy = async (text: string, message: string) => {
  try {
    await writeToClipboard(text)
    snackbar.showSnackbar(message, 'default', 2000)
  } catch {
    snackbar.showSnackbar('Unable to copy to clipboard', 'error', 2000)
  }
}

const copyLink = () => copy(pack.value!.packTomlUrl, 'Link copied to clipboard')
const copyInstanceUrl = () => copy(pack.value!.multimcUrl, 'Import URL copied to clipboard')
const copySetup = () => copy(
  getClientSetupCommand(pack.value!.packTomlUrl),
  'Client setup command copied to clipboard',
)

const downloading = ref(false)
const downloadInstance = async () => {
  downloading.value = true
  try {
    await downloadInstanceZip(pack.value!.multimcUrl)
  } catch (e) {
    snackbar.showSnackbar((e as Error).message, 'error', 3000)
  } finally {
    downloading.value = false
  }
}

const defaultTitle = document.title

watchEffect(() => {
  document.title = pack.value ? pack.value.name : defaultTitle
})

onUnmounted(() => {
  document.title = defaultTitle
})
</script>

<template>
  <v-container
    class="py-8"
    max-width="960"
  >
    <div
      v-if="isLoading"
      class="d-flex justify-center py-12"
    >
      <v-progress-circular indeterminate />
    </div>

    <v-alert
      v-else-if="notFound"
      type="info"
      variant="tonal"
      title="Pack not found"
      text="This pack does not exist or is not public."
    />

    <v-alert
      v-else-if="error || !pack"
      type="error"
      variant="tonal"
      title="Could not load pack"
      text="Try again in a moment."
    />

    <template v-else>
      <v-sheet
        color="primary"
        class="pa-6 mb-4"
        elevation="4"
      >
        <div class="d-flex align-center ga-4 mb-2">
          <v-img
            src="/android-chrome-192x192.png"
            alt="Packwiz Web"
            width="64"
            max-width="64"
            height="64"
          />
          <div>
            <h1>{{ pack.name }}</h1>
            <div class="text-body-2">
              by {{ pack.author }}
            </div>
          </div>
        </div>
        <p
          v-if="pack.description"
          class="mb-3 text-pre-wrap"
        >
          {{ pack.description }}
        </p>
        <div class="d-flex flex-wrap ga-2">
          <v-chip
            v-if="pack.mcVersion"
            prepend-icon="mdi-minecraft"
            :text="pack.mcVersion"
            variant="flat"
            label
          />
          <v-chip
            v-if="pack.loader"
            prepend-icon="mdi-puzzle"
            :text="`${pack.loader} ${pack.loaderVersion}`.trim()"
            variant="flat"
            label
          />
          <v-chip
            v-if="pack.version"
            prepend-icon="mdi-tag-outline"
            :text="`v${pack.version}`"
            variant="flat"
            label
          />
          <v-chip
            prepend-icon="mdi-package-variant"
            :text="`${pack.mods.length} mods`"
            variant="flat"
            label
          />
        </div>
      </v-sheet>

      <v-card class="mb-4">
        <v-card-title class="text-primary">
          <v-icon
            icon="mdi-download"
            class="me-2"
          />
          Install
        </v-card-title>
        <v-card-text>
          <h4 class="mb-1">
            For MultiMC or Prism Launcher
          </h4>
          <p class="text-medium-emphasis mb-3">
            Download the instance zip, or copy its URL, then use "Add Instance" &rarr; "Import"
            in the launcher. It is set up with the right Minecraft and loader version and keeps
            itself in sync with this pack.
          </p>
          <div class="d-flex flex-wrap ga-2 mb-6">
            <v-btn
              color="primary"
              prepend-icon="mdi-download"
              text="Download instance (.zip)"
              :loading="downloading"
              @click="downloadInstance"
            />
            <v-btn
              variant="tonal"
              prepend-icon="mdi-clipboard-text-multiple-outline"
              text="Copy import URL"
              @click="copyInstanceUrl"
            />
          </div>

          <h4 class="mb-1">
            Other launchers
          </h4>
          <p class="text-medium-emphasis mb-3">
            Use the pack link with any packwiz installer. To set up an instance by hand, paste the
            setup command as its pre-launch command.
          </p>
          <v-text-field
            :model-value="pack.packTomlUrl"
            prepend-inner-icon="mdi-link-variant"
            label="pack.toml"
            readonly
            hide-details
            density="comfortable"
            class="mb-3"
          />
          <div class="d-flex flex-wrap ga-2">
            <v-btn
              color="primary"
              prepend-icon="mdi-clipboard-text-multiple-outline"
              text="Copy pack.toml link"
              @click="copyLink"
            />
            <v-btn
              variant="tonal"
              prepend-icon="mdi-console"
              text="Copy client setup command"
              @click="copySetup"
            />
          </div>
          <a
            class="d-inline-block mt-3"
            href="https://packwiz.infra.link/tutorials/installing/packwiz-installer/"
            target="_blank"
            rel="noopener noreferrer"
          >
            packwiz-installer guide
          </a>
        </v-card-text>
      </v-card>

      <v-card class="mb-4">
        <v-card-title class="text-primary">
          <v-icon
            icon="mdi-information-outline"
            class="me-2"
          />
          Details
        </v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col
              v-for="d in details"
              :key="d.label"
              cols="12"
              sm="6"
              md="4"
            >
              <div class="text-caption text-medium-emphasis">
                {{ d.label }}
              </div>
              <div>{{ d.value }}</div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card>
        <v-card-title class="text-primary">
          <v-icon
            icon="mdi-package-variant"
            class="me-2"
          />
          Mods ({{ pack.mods.length }})
        </v-card-title>
        <v-card-text>
          <div class="d-flex flex-wrap ga-4 align-center mb-2">
            <v-text-field
              v-model="search"
              label="Search mods"
              prepend-inner-icon="mdi-magnify"
              clearable
              hide-details
              density="compact"
              max-width="320"
            />
            <v-checkbox
              v-model="hideDependencies"
              label="Hide dependencies"
              hide-details
              density="compact"
            />
          </div>
          <v-data-table
            :headers="headers"
            :items="mods"
            density="comfortable"
            items-per-page="25"
          >
            <template #[`item.name`]="{ item }">
              {{ item.name }}
              <v-chip
                v-if="item.optional"
                size="small"
                color="info"
                variant="tonal"
                label
                class="ms-1"
              >
                optional
              </v-chip>
            </template>
          </v-data-table>
        </v-card-text>
      </v-card>
    </template>
  </v-container>
</template>
