<route lang="yaml">
meta:
  layout: app
</route>

<script lang="ts" setup>
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {useBulkAdd} from "@/composables/useBulkAdd.ts";
import type {Pack} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import {getCurseforgeStatus} from "@/services/mods.service.ts";
import {isActiveBulkStatus, progressText} from "@/lib/bulk-add.ts";
import BulkAddForm from "@/components/mods/BulkAddForm.vue";
import BulkAddResults from "@/components/mods/BulkAddResults.vue";

const route = useRoute<'/packs/[packId].bulk-add'>()
const router = useRouter()
const packId = () => Number(route.params.packId)

const {isLoading, data: pack} = buildDataLoader<Pack>(async () => fetchOnePack(packId()))

const {run, loading: runLoading, starting, error, start, updateItem, clear} = useBulkAdd(packId)

const text = ref("")
const curseforgeAvailable = ref(false)

onMounted(async () => {
  try {
    curseforgeAvailable.value = (await getCurseforgeStatus(packId())).available
  } catch (e) {
    console.error("Failed to check CurseForge status:", e)
  }
})

const running = computed(() => isActiveBulkStatus(run.value?.status))

const submit = async (value: string) => {
  if (await start(value)) text.value = ""
}

const backToPack = () => router.push({path: `/packs/${packId()}`})
</script>

<template>
  <div class="ma-6">
    <v-skeleton-loader
      v-if="isLoading || runLoading"
      elevation="0"
      type="heading, subtitle, paragraph@2"
    />

    <v-card v-else-if="pack">
      <v-card-title class="d-flex align-center">
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
        <v-spacer />
        <v-btn
          text="Back to pack"
          prepend-icon="mdi-arrow-left"
          variant="text"
          @click="backToPack"
        />
      </v-card-title>

      <v-card-subtitle class="d-flex align-center">
        <h3>Add Mods from a List</h3>
        <v-spacer />
        <v-btn
          v-if="run && !running"
          text="New list"
          prepend-icon="mdi-playlist-plus"
          variant="tonal"
          @click="clear"
        />
      </v-card-subtitle>

      <div class="ma-6">
        <v-alert
          v-if="error"
          type="error"
          variant="tonal"
          density="compact"
          class="mb-4"
          :text="error"
        />

        <BulkAddForm
          v-if="!run"
          v-model="text"
          :starting="starting"
          @submit="submit"
        />

        <div v-else-if="running">
          <p class="mb-2">
            {{ progressText(run.status, run.items.length) }}
          </p>
          <v-progress-linear
            color="primary"
            indeterminate
            class="mb-2"
          />
          <p class="text-medium-emphasis">
            You can leave this page; the results will be here when you come back.
          </p>
        </div>

        <template v-else>
          <v-alert
            v-if="run.status === 'failed'"
            type="error"
            variant="tonal"
            density="compact"
            class="mb-4"
            :text="`The bulk add failed${run.error ? `: ${run.error}` : ''}`"
          />
          <BulkAddResults
            :pack="pack"
            :run="run"
            :curseforge-available="curseforgeAvailable"
            @resolved="updateItem"
          />
        </template>
      </div>
    </v-card>
  </div>
</template>
