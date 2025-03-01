<script setup lang="ts">
import {buildDataLoader} from "@/composables/data-loader.ts";
import {type Pack} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import PackActions from "@/components/PackActions.vue";
import SearchBar from "@/components/SearchBar.vue";


const {slug} = defineProps<{ slug: string }>()
const search = ref('')

const {
  isLoading,
  data: pack,
  reload,
  error,
} = buildDataLoader<Pack>(async () => {
  return fetchOnePack(slug)
})
</script>

<template>
  <div
    v-if="isLoading"
    class="ma-6"
  >
    <v-skeleton-loader
      elevation="0"
      theme="article"
      type="heading, subtitle, actions, paragraph@2"/>
  </div>

  <div
    v-else-if="pack"
    class="ma-6"
  >
    <v-card-title
      class="d-flex align-center"
    >
      <template
        class="d-flex align-center me-auto"
      >
        <h1 class="me-5">{{ pack.title }}</h1>
        <PackStatus :status="pack.archived ? 'archived' : pack.status"/>
        <div
          class="ms-2"
          v-if="pack.isPublic">
          <PackStatus status="public"/>
        </div>
      </template>

      <template v-if="!pack.archived" class="d-flex align-center">
        <v-btn
          v-if="pack.status === 'draft'"
          text="Publish"
        />
        <v-btn
          v-else-if="pack.status === 'published'"
          text="Convert to Draft"
          color="warning"
        />
      </template>

      <v-btn
        v-if="!pack.archived"
        class="ms-3"
        text="Archive"
        color="error"
      />
      <v-btn
        v-else
        class="ms-3"
        text="Unarchive"
        color="error"
      />

      <PackActions class="ms-3" :pack="pack"/>
      <v-btn
        icon="mdi-refresh"
        variant="text"
        color="disabled"
        @click="reload"
      />
    </v-card-title>

    <v-divider/>

    <v-card-text class="ma-2">
      <h3 v-if="pack.packData && pack.slug != pack.packData?.name" class="mb-4">
        Slug: {{ pack.slug }}
      </h3>
      <p>
        {{ pack.description }}
      </p>
    </v-card-text>

    <v-toolbar>
      <v-toolbar-title>Mods</v-toolbar-title>


      <SearchBar
        max-width="400"
        class="me-2"
        v-model="search"
        density="comfortable"
      />
    </v-toolbar>
  </div>
</template>
