<script setup lang="ts">
import type {Mod, ModDependency, ModSearchResult} from "@/interfaces/pack.ts";
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import {getResultState, modPageUrl} from "@/lib/mod-source.ts";
import {formatAuthorLine, isPackLoader, limitChips, prettyCategory} from "@/lib/search-meta.ts";

const {results, source, installedMods, addedKeys, addingSlug, selectedSlug, dependencies, errorMessage, dependenciesFailed, packLoader = ""} = defineProps<{
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

const MAX_LOADERS = 3
const MAX_CATEGORIES = 2

// Per-row display data, computed once per results change.
const rows = computed(() => results.map(result => {
  const allLoaders = result.loaders ?? []
  const allCategories = source === "modrinth"
    ? (result.categories ?? []).map(prettyCategory)
    : result.categories ?? []
  return {
    result,
    authorLine: formatAuthorLine(result.author, result.downloads),
    loaders: limitChips(allLoaders, MAX_LOADERS).shown,
    hiddenLoaders: allLoaders.slice(MAX_LOADERS),
    categories: limitChips(allCategories, MAX_CATEGORIES).shown,
    hiddenCategories: allCategories.slice(MAX_CATEGORIES),
  }
}))

const addLabel = computed(() =>
  dependencies.length > 0 ? "Add Mod and Dependencies" : "Add Mod"
)
</script>

<template>
  <v-list
    max-height="450"
    class="overflow-y-auto mb-4 pa-0 bg-transparent"
  >
    <template
      v-for="{result, authorLine, loaders, hiddenLoaders, categories, hiddenCategories} in rows"
      :key="result.projectId"
    >
      <v-list-item
        :active="result.slug === selectedSlug"
        :link="isAvailable(result)"
        color="primary"
        rounded="lg"
        class="mb-2 elevation-2 bg-surface"
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
          v-if="authorLine"
          class="text-truncate"
        >
          {{ authorLine }}
        </v-list-item-subtitle>
        <div
          v-if="loaders.length > 0 || categories.length > 0"
          class="d-flex flex-wrap ga-1 mt-1"
          style="min-width: 0"
        >
          <div
            v-if="loaders.length > 0"
            role="list"
            aria-label="Loaders"
            class="d-flex flex-wrap ga-1"
            style="min-width: 0"
          >
            <v-chip
              v-for="loader in loaders"
              :key="`l-${loader}`"
              role="listitem"
              size="x-small"
              label
              :color="isPackLoader(loader, packLoader) ? 'primary' : undefined"
              :variant="isPackLoader(loader, packLoader) ? 'flat' : 'tonal'"
              :text="loader"
            />
            <v-chip
              v-if="hiddenLoaders.length > 0"
              v-tooltip="hiddenLoaders.join(', ')"
              role="listitem"
              size="x-small"
              label
              variant="outlined"
              :aria-label="`${hiddenLoaders.length} more loaders: ${hiddenLoaders.join(', ')}`"
              :text="`+${hiddenLoaders.length}`"
            />
          </div>
          <div
            v-if="categories.length > 0"
            role="list"
            aria-label="Categories"
            class="d-flex flex-wrap ga-1"
            style="min-width: 0"
          >
            <v-chip
              v-for="category in categories"
              :key="`c-${category}`"
              role="listitem"
              size="x-small"
              variant="tonal"
              style="max-width: 10rem"
              :title="category"
              :text="category"
            />
            <v-chip
              v-if="hiddenCategories.length > 0"
              v-tooltip="hiddenCategories.join(', ')"
              role="listitem"
              size="x-small"
              variant="outlined"
              :aria-label="`${hiddenCategories.length} more categories: ${hiddenCategories.join(', ')}`"
              :text="`+${hiddenCategories.length}`"
            />
          </div>
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
        class="px-4 pb-4 mb-2"
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
