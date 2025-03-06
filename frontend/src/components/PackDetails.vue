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
    text: `Version: ${ pack.packData?.version }`,
    color: "teal",
    condition: !!pack.packData,
  },
  {
    text: `Minecraft: ${ pack.packData?.versions.minecraft }`,
    color: "cyan",
    condition: !!pack.packData,
  },
  {
    text: `${toTitleCase(pack.packData?.versions.loader.type || "")}: ${ pack.packData?.versions.loader.version }`,
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
          <PackStatus :status="pack.isArchived ? 'archived' : pack.status" />
          <div
            v-if="pack.isPublic"
            class="ms-2"
          >
            <PackStatus status="public" />
          </div>
        </div>

        <v-btn
          v-if="pack.permission >= PackPermission.EDIT || authStore.user?.isAdmin"
          prepend-icon="mdi-pencil"
          text="Edit"
          :to="`/packs/${pack.slug}/edit`"
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

        <PackActions
          class="ms-3"
          :pack="pack"
        />
        <v-btn
          icon="mdi-refresh"
          variant="text"
          color="disabled"
          @click="$emit('reload')"
        />
      </v-card-title>

      <v-divider />

      <v-card-text class="ma-2">
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
          class="mt-6"
        >
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
