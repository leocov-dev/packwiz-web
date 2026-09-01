<script setup lang="ts">
import type {PackResponse, MigrateDryRunResponse} from "@/interfaces/pack.ts"
import type {MigratePackRequest} from "@/interfaces/requests.ts"
import type {LoaderVersions} from "@/stores/cache.ts"
import MinecraftVersion, {LATEST_SENTINEL, LATEST_SNAPSHOT_SENTINEL} from "@/components/forms/MinecraftVersion.vue"
import Loader from "@/components/forms/Loader.vue"
import {migratePack, migrateDryRun} from "@/services/packs.service.ts"
import {useSnackbarStore} from "@/stores/snackbar.ts"

const {pack} = defineProps<{ pack: PackResponse }>()
const model = defineModel<boolean>({required: true})
const emit = defineEmits(["migrated"])

const snackbar = useSnackbarStore()

const minecraftVersion = ref(pack.mcVersion)
const loader = ref<{ name?: keyof LoaderVersions, version?: string }>({
  name: pack.loader as keyof LoaderVersions,
  version: pack.loaderVersion,
})
const updateMods = ref(true)
const useRecommended = ref(false)

const isValid = ref(false)
const loading = ref(false)

const dryRunResult = ref<MigrateDryRunResponse | null>(null)
const dryRunLoading = ref(false)
const dryRunError = ref(false)

const isForge = computed(() => (loader.value.name || "").toLowerCase() === "forge")

const dryRunSummary = computed(() => {
  if (!dryRunResult.value) return null
  const mods = dryRunResult.value.mods
  return {
    updatable: mods.filter(m => m.updateAvailable).length,
    incompatible: mods.filter(m => m.incompatible).length,
  }
})

const buildRequest = (): MigratePackRequest => {
  const isLatest = minecraftVersion.value === LATEST_SENTINEL
  const isSnapshot = minecraftVersion.value === LATEST_SNAPSHOT_SENTINEL

  return {
    minecraft: {
      version: (isLatest || isSnapshot) ? "" : minecraftVersion.value || "",
      latest: isLatest,
      snapshot: isSnapshot,
    },
    loader: {
      name: (loader.value.name || "").toLowerCase(),
      version: loader.value.version || "",
      latest: false,
    },
    updateMods: updateMods.value,
    useRecommended: isForge.value && useRecommended.value,
  }
}

const preview = async () => {
  dryRunLoading.value = true
  dryRunError.value = false
  try {
    dryRunResult.value = await migrateDryRun(pack.id, buildRequest())
  } catch (e) {
    console.error(e)
    dryRunError.value = true
  } finally {
    dryRunLoading.value = false
  }
}

// invalidate a stale preview whenever the target changes
watch([minecraftVersion, () => loader.value.name, () => loader.value.version, useRecommended], () => {
  dryRunResult.value = null
  dryRunError.value = false
})

const submit = async () => {
  loading.value = true
  try {
    await migratePack(pack.id, buildRequest())
    model.value = false
    emit("migrated")
  } catch (e) {
    console.error(e)
    snackbar.showSnackbar("Failed to migrate pack", "error", 4000)
  } finally {
    loading.value = false
  }
}

const cancel = () => {
  model.value = false
}
</script>

<template>
  <v-dialog
    v-model="model"
    persistent
    max-width="600"
  >
    <v-card class="pa-3">
      <v-card-title>Migrate Pack</v-card-title>
      <v-card-text>
        <v-form
          v-model="isValid"
          validate-on="eager"
        >
          <MinecraftVersion
            v-model:version="minecraftVersion"
            :include-latest="true"
          />
          <Loader
            v-model:loader="loader.name"
            v-model:version="loader.version"
            :minecraft-version="minecraftVersion || ''"
          />
          <v-checkbox
            v-if="isForge"
            v-model="useRecommended"
            label="Use Forge's recommended build for this Minecraft version (overrides the loader version above)"
            hide-details
          />
          <v-checkbox
            v-model="updateMods"
            label="Also update all mods to match the new Minecraft version / loader"
            hide-details
          />
        </v-form>

        <v-btn
          class="mt-2"
          variant="outlined"
          size="small"
          :loading="dryRunLoading"
          :disabled="!isValid"
          @click="preview"
        >
          Preview
        </v-btn>

        <v-alert
          v-if="dryRunError"
          class="mt-3"
          type="error"
          density="compact"
          text="Failed to preview this migration"
        />

        <div
          v-if="dryRunSummary"
          class="mt-3"
        >
          <div class="text-body-2 mb-2">
            {{ dryRunSummary.updatable }} mod(s) will update,
            {{ dryRunSummary.incompatible }} mod(s) need attention
          </div>
          <v-list density="compact">
            <v-list-item
              v-for="mod in dryRunResult!.mods.filter(m => m.incompatible || (m.pinned && m.updateAvailable))"
              :key="mod.modId"
            >
              <template #prepend>
                <v-icon
                  v-if="mod.incompatible"
                  v-tooltip="mod.error"
                  color="error"
                  icon="mdi-alert-circle"
                  class="me-2"
                />
                <v-icon
                  v-else
                  v-tooltip="'pinned, update available but will be skipped'"
                  color="warning"
                  icon="mdi-pin"
                  class="me-2"
                />
              </template>
              <v-list-item-title>{{ mod.name }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </div>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          class="me-3"
          color="primary"
          variant="tonal"
          :disabled="loading"
          @click="cancel"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          :loading="loading"
          :disabled="loading || !isValid"
          @click="submit"
        >
          Migrate
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
