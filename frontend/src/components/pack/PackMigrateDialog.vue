<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts"
import type {MigratePackRequest} from "@/interfaces/requests.ts"
import type {LoaderVersions} from "@/stores/cache.ts"
import MinecraftVersion, {LATEST_SENTINEL, LATEST_SNAPSHOT_SENTINEL} from "@/components/forms/MinecraftVersion.vue"
import Loader from "@/components/forms/Loader.vue"
import {migratePack} from "@/services/packs.service.ts"
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

const isForge = computed(() => (loader.value.name || "").toLowerCase() === "forge")

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
