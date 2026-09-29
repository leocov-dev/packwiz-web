<script setup lang="ts">
import type {Mod, ModDependency, ModSearchResult} from "@/interfaces/pack.ts";
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import {getResultState, modPageUrl} from "@/lib/mod-source.ts";
import {formatAuthorLine, isPackLoader, limitChips} from "@/lib/search-meta.ts";

const {results, source, installedMods, addedKeys, addingSlug, selectedSlug, dependencies, errorMessage, dependenciesFailed, packLoader} = defineProps<{
  results: ModSearchResult[]
  source: "modrinth" | "curseforge"
  installedMods: Mod[]
  addedKeys: string[]
  addingSlug: string
  selectedSlug: string
  dependencies: ModDependency[]
  errorMessage: string
  dependenciesFailed: boolean
  packLoader?: string
}>()

const emit = defineEmits<{
  select: [result: ModSearchResult]
  add: [result: ModSearchResult]
  retryDependencies: []
}>()

const stateOf = (result: ModSearchResult) => getResultState(result, installedMods, addedKeys, source)

const onRowClick = (result: ModSearchResult) => {
  if (isAvailable(result)) {
    emit("select", result)
  }
}

const isAvailable = (result: ModSearchResult) => stateOf(result) === "available"

const loaderChips = (result: ModSearchResult) => limitChips(result.loaders, 3).shown
const categoryChips = (result: ModSearchResult) => limitChips(result.categories, 2).shown

const addLabel = computed(() =>
  dependencies.length > 0 ? "Add Mod and Dependencies" : "Add Mod"
)
</script>

<template>
  <v-list
    max-height="450"
    class="overflow-y-auto mb-4"
  >
    <template
      v-for="result in results"
      :key="result.projectId"
    >
      <v-list-item
        :active="result.slug === selectedSlug"
        :link="isAvailable(result)"
        @click="onRowClick(result)"
      >
        <template #prepend>
          <v-avatar
            v-if="result.iconUrl"
            :image="result.iconUrl"
          />
          <v-icon
            v-else
            icon="mdi-puzzle-outline"
          />
        </template>
        <v-list-item-title>{{ result.title }}</v-list-item-title>
        <v-list-item-subtitle class="text-truncate">
          {{ result.description }}
        </v-list-item-subtitle>
        <v-list-item-subtitle
          v-if="formatAuthorLine(result.author, result.downloads)"
          class="text-truncate"
        >
          {{ formatAuthorLine(result.author, result.downloads) }}
        </v-list-item-subtitle>
        <div
          v-if="loaderChips(result).length > 0 || categoryChips(result).length > 0"
          class="d-flex flex-nowrap ga-1 mt-1 overflow-hidden"
        >
          <v-chip
            v-for="loader in loaderChips(result)"
            :key="`l-${loader}`"
            size="x-small"
            label
            class="flex-shrink-0"
            :color="isPackLoader(loader, packLoader) ? 'primary' : undefined"
            :variant="isPackLoader(loader, packLoader) ? 'flat' : 'tonal'"
            :text="loader"
          />
          <v-chip
            v-for="category in categoryChips(result)"
            :key="`c-${category}`"
            size="x-small"
            variant="tonal"
            class="text-truncate"
            :text="category"
          />
        </div>
        <template #append>
          <v-chip
            v-if="stateOf(result) === 'installed'"
            color="success"
            size="small"
            prepend-icon="mdi-check-circle"
            text="Already installed"
          />
          <v-chip
            v-else-if="stateOf(result) === 'added'"
            color="success"
            size="small"
            prepend-icon="mdi-check-circle"
            text="Added"
          />
          <v-btn
            v-tooltip="'Open in new tab'"
            :aria-label="`Open ${result.title} in new tab`"
            icon="mdi-open-in-new"
            variant="text"
            size="small"
            density="comfortable"
            class="ms-2"
            :href="modPageUrl(source, result.slug)"
            target="_blank"
            rel="noopener"
            @click.stop
          />
        </template>
      </v-list-item>

      <div
        v-if="result.slug === selectedSlug && isAvailable(result)"
        class="px-4 pb-4"
      >
        <MissingDependencies
          v-if="dependencies.length > 0"
          class="mb-3"
          :missing="dependencies"
        />
        <v-alert
          v-if="dependenciesFailed"
          type="warning"
          variant="tonal"
          density="compact"
          class="mb-3"
          text="Couldn't check dependencies. You can still add this mod, but required dependencies may be missing."
        >
          <template #append>
            <v-btn
              text="Retry"
              variant="text"
              size="small"
              @click="emit('retryDependencies')"
            />
          </template>
        </v-alert>
        <v-alert
          v-if="errorMessage"
          type="error"
          variant="tonal"
          density="compact"
          class="mb-3"
          :text="errorMessage"
        />
        <div class="d-flex justify-end">
          <v-btn
            :text="addLabel"
            color="primary"
            :loading="addingSlug === result.slug"
            :disabled="addingSlug !== ''"
            @click="emit('add', result)"
          />
        </div>
      </div>
    </template>
  </v-list>
</template>
