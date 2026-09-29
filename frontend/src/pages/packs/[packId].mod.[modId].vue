<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {type Pack, type Mod} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import {fetchOneMod} from "@/services/mods.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {describeUpdateResult, snapshotMod, type ModSnapshot} from "@/lib/mod-edit.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const route = useRoute<'/packs/[packId].mod.[modId]'>()

const {
  isLoading,
  data,
  reload,
  error,
} = buildDataLoader<{ pack: Pack, mod: Mod }>(async () => {
  const pack = await fetchOnePack(Number(route.params.packId), true)
  const mod = await fetchOneMod(Number(route.params.packId), Number(route.params.modId))
  return {pack, mod}
})

const snackbar = useSnackbarStore()

// Reload after an update-from-source, then report what actually changed.
const onReload = async (before: ModSnapshot) => {
  await reload()
  const mod = data.value?.mod
  if (error.value || !mod) {
    snackbar.showSnackbar(error.value ? apiErrorMessage(error.value, "Updated, but failed to reload the mod") : "Updated, but failed to reload the mod", "error")
    return
  }
  snackbar.showSnackbar(describeUpdateResult(mod.name || mod.slug, before, snapshotMod(mod)), "success")
}
</script>

<template>
  <div
    v-if="isLoading"
    class="ma-6"
  >
    <v-skeleton-loader
      elevation="0"
      theme="article"
      type="heading, subtitle, actions, paragraph@2"
    />
  </div>

  <v-alert
    v-else-if="error"
    class="ma-6"
    type="error"
    icon="mdi-alert"
    :text="apiErrorMessage(error, 'Failed to load mod')"
  />

  <EditModForm
    v-else-if="data?.pack && data?.mod"
    :pack="data.pack"
    :mod="data.mod"
    @reload="onReload"
  />
</template>

