<script setup lang="ts">
import {buildDataLoader} from "@/composables/data-loader.ts";
import {type Pack, PackPermission} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import PackActions from "@/components/PackActions.vue";
import ModsList from "@/components/ModsList.vue";
import {toTitleCase} from "../services/utils.ts";
import {useAuthStore} from "@/stores/auth.ts";


const {slug} = defineProps<{ slug: string }>()
const search = ref('')

const authStore = useAuthStore()

const {
  isLoading,
  data: pack,
  reload,
  error,
} = buildDataLoader<Pack>(async () => {
  return fetchOnePack(slug)
})

const onAddMod = () => {
  alert('add mod')
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
      type="heading, subtitle, actions, paragraph@2"/>
  </div>

  <div
    v-else-if="pack"
    class="ma-6"
  >
    <v-card>
      <v-card-title
        class="d-flex align-center"
      >
        <template
          class="d-flex align-center me-auto"
        >
          <h1 class="me-5">{{ pack.title }}</h1>
          <PackStatus :status="pack.isArchived ? 'archived' : pack.status"/>
          <div
            class="ms-2"
            v-if="pack.isPublic">
            <PackStatus status="public"/>
          </div>
        </template>

        <v-btn
          v-if="pack.permission >= PackPermission.EDIT || authStore.user?.isAdmin"
          prepend-icon="mdi-pencil"
          text="Edit"
        />

<!--        <template v-if="!pack.isArchived" class="d-flex align-center">-->
<!--          <v-btn-->
<!--            v-if="pack.status === 'draft'"-->
<!--            text="Publish"-->
<!--          />-->
<!--          <v-btn-->
<!--            v-else-if="pack.status === 'published'"-->
<!--            text="Convert to Draft"-->
<!--            color="warning"-->
<!--          />-->
<!--        </template>-->

<!--        <v-btn-->
<!--          v-if="!pack.isArchived"-->
<!--          class="ms-3"-->
<!--          text="Archive"-->
<!--          color="error"-->
<!--        />-->
<!--        <v-btn-->
<!--          v-else-->
<!--          class="ms-3"-->
<!--          text="Unarchive"-->
<!--          color="error"-->
<!--        />-->

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
        <template v-if="pack.packData">
          <div class="d-flex mb-3">
            <v-chip
              v-if="pack.slug != pack.packData?.name"
              class="me-2"
              color="orange"
            >
              {{ pack.slug }}
            </v-chip>
            <v-chip
              class="me-2"
              color="teal"
            >
              Version: {{ pack.packData.version }}
            </v-chip>
            <v-chip
              class="me-2"
              color="cyan"
            >
               Minecraft: {{ pack.packData.versions.minecraft }}
            </v-chip>
            <v-chip
              class="me-2"
              color="yellow"
            >
              {{ toTitleCase(pack.packData.versions.loader.type) }}: {{ pack.packData.versions.loader.version }}
            </v-chip>
            <v-chip
              class="me-2"
              color="purple"
            >
              {{ pack.packData.packFormat }}
            </v-chip>
          </div>
        </template>
        <p v-if="pack.description" class="mt-6">
          {{ pack.description }}
        </p>
      </v-card-text>
    </v-card>

    <ModsList
      :mods="pack.modData || []"
      :can-edit="pack.permission >= PackPermission.EDIT"
      @add-mod="onAddMod"
    />
  </div>
</template>
