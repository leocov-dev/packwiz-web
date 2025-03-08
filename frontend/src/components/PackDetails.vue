<script setup lang="ts">
import {type Pack, PackPermission} from "@/interfaces/pack.ts";
import PackActions from "@/components/PackActions.vue";
import ModsList from "@/components/ModsList.vue";
import {toTitleCase} from "../services/utils.ts";
import {useAuthStore} from "@/stores/auth.ts";

const {pack} = defineProps<{ pack: Pack }>()

defineEmits(['reload'])

const router = useRouter()
const authStore = useAuthStore()

const onAddMod = () => {
  router.push({path: `/packs/${pack.slug}/add-mod`})
}

interface Chip {
  text: string,
  color: string,
  condition: boolean,
}

const chipList: Chip[] = [
  {
    text: pack.slug,
    color: "orange",
    condition: pack.slug != pack.packData?.name,
  },
  {
    text: `Version: ${pack.packData?.version}`,
    color: "teal",
    condition: !!pack.packData,
  },
  {
    text: `Minecraft: ${pack.packData?.versions.minecraft}`,
    color: "cyan",
    condition: !!pack.packData,
  },
  {
    text: `${toTitleCase(pack.packData?.versions.loader.type || "")}: ${pack.packData?.versions.loader.version}`,
    color: "yellow",
    condition: !!pack.packData,
  },
  {
    text: pack.packData?.packFormat || "",
    color: "purple",
    condition: !!pack.packData,
  },
  {
    text: `Game Versions: ${pack.packData?.options?.acceptableGameVersions?.join(", ") || ""}`,
    color: "magenta",
    condition: !!pack.packData?.options?.acceptableGameVersions?.length,
  },
].filter(chip => chip.condition)

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>
      <v-card-title
        class="d-flex flex-wrap align-center"
      >
        <div
          class="d-flex align-center me-auto"
        >
          <h1 class="me-5">
            {{ pack.title }}
          </h1>
          <PackStatus :status="pack.isArchived ? 'archived' : pack.status"/>
          <PackStatus v-if="pack.isPublic" status="public"/>
          <PackStatus v-if="pack.dataMissing" status="warning"/>
        </div>

        <PackActions
          class="ms-3"
          :pack="pack"
        />
        <!--        <v-btn-->
        <!--          icon="mdi-refresh"-->
        <!--          variant="text"-->
        <!--          color="disabled"-->
        <!--          @click="$emit('reload')"-->
        <!--        />-->
      </v-card-title>

      <v-divider/>

      <div
        v-if="!pack.dataMissing"
        class="d-flex align-center justify-end mt-3 ms-3 me-3">
        <v-btn
          v-if="pack.permission >= PackPermission.EDIT || authStore.user?.isAdmin"
          prepend-icon="mdi-pencil"
          text="Edit"
          :to="`/packs/${pack.slug}/edit`"
          class="me-3"
        />

        <div
          v-if="!pack.isArchived"
          class="d-flex align-center me-3"
        >
          <v-btn
            v-if="pack.status === 'draft'"
            text="Publish"
            prepend-icon="mdi-earth"
          />
          <v-btn
            v-else-if="pack.status === 'published'"
            text="Convert to Draft"
            color="warning"
            prepend-icon="mdi-file-edit"
          />
        </div>

        <v-btn
          v-if="!pack.isArchived"
          text="Archive"
          color="error"
          prepend-icon="mdi-archive"
        />
        <v-btn
          v-else
          class="ms-3"
          text="Unarchive"
          color="error"
          prepend-icon="mdi-archive-refresh"
        />
      </div>
      <div v-else
        class="ma-4"
      >
        <v-alert
          icon="mdi-alert"
          type="warning"
          text="This pack has invalid or missing file data."
        />
      </div>

      <v-card-text class="ms-2 me-2 mb-2">
        <div
          v-if="chipList.length > 0"
          class="d-flex flex-wrap align-center"
        >
          <v-chip
            v-for="chip in chipList"
            :key="chip.text"
            class="me-2 mb-2"
            :color="chip.color"
          >
            {{ chip.text }}
          </v-chip>
        </div>
        <p
          v-if="pack.description"
          class="mt-6 text-pre-line"
        >
          {{ pack.description }}
        </p>
      </v-card-text>
    </v-card>

    <ModsList
      :mods="pack.modData || []"
      :can-edit="pack.permission >= PackPermission.EDIT"
      :disabled="pack.dataMissing"
      @add-mod="onAddMod"
    />
  </div>
</template>
