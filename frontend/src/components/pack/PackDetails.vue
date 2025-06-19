<script setup lang="ts">
import {PackPermission, PackResponse} from "@/interfaces/pack.ts";
import PackActions from "@/components/pack/PackActions.vue";
import ModsList from "@/components/mods/ModsList.vue";
import {toTitleCase} from "@/services/utils.ts";
import {useAuthStore} from "@/stores/auth.ts";
import {
  archivePack,
  convertPackToDraft,
  makePackPrivate,
  makePackPublic,
  publishPack,
  unArchivePack
} from "@/services/packs.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";

const {pack} = defineProps<{ pack: PackResponse }>()

defineEmits(['reload'])

const showPublishDialog = ref(false)
const showDraftDialog = ref(false)
const showArchiveDialog = ref(false)
const showUnArchiveDialog = ref(false)
const showPublicDialog = ref(false)
const showPrivateDialog = ref(false)

const router = useRouter()
const authStore = useAuthStore()

const onAddMod = () => {
  router.push({path: `/packs/${pack.id}/add-mod`})
}

interface Chip {
  text: string,
  color: string,
}

const chipList: Chip[] = [
  {
    text: pack.slug,
    color: "orange",
  },
  {
    text: `Version: ${pack.version}`,
    color: "teal",
  },
  {
    text: `Minecraft: ${pack.mcVersion}`,
    color: "cyan",
  },
  {
    text: `${toTitleCase(pack.loader)}: ${pack.loaderVersion}`,
    color: "yellow",
  },
  {
    text: pack.packFormat,
    color: "purple",
  },
  {
    text: `Game Versions: ${(pack.acceptableGameVersions || [pack.mcVersion]).join(', ')}`,
    color: "magenta",
  },
]


const convertToDraft = async () => {
  await convertPackToDraft(pack.id)
  router.go(0)
}

const publish = async () => {
  await publishPack(pack.id)
  router.go(0)
}

const archive = async () => {
  await archivePack(pack.id)
  router.go(0)
}

const unArchive = async () => {
  await unArchivePack(pack.id)
  router.go(0)
}

const makePublic = async () => {
  await makePackPublic(pack.id)
  router.go(0)
}

const makePrivate = async () => {
  await makePackPrivate(pack.id)
  router.go(0)
}

</script>

<template>
  <div
    class="ma-6"
  >
    <ConfirmationDialog
      v-model="showPublishDialog"
      title="Confirm Publish Pack"
      text="Are you sure you want to publish this pack?
      It will be accessible to users."
      @accepted="publish"
    />
    <ConfirmationDialog
      v-model="showDraftDialog"
      title="Confirm Convert to Draft"
      text="Are you sure you want to convert this pack to draft?
      Users will not be able to access this pack."
      @accepted="convertToDraft"
    />
    <ConfirmationDialog
      v-model="showArchiveDialog"
      title="Confirm Archive Pack"
      text="Are you sure you want to archive this pack?
      Users will not be able to access this pack and you won't be able to edit it."
      @accepted="archive"
    />
    <ConfirmationDialog
      v-model="showUnArchiveDialog"
      title="Confirm Unarchive Pack"
      text="Are you sure you want to unarchive this pack?"
      @accepted="unArchive"
    />
    <ConfirmationDialog
      v-model="showPublicDialog"
      title="Confirm Make Pack Public"
      text="Are you sure you want to make this pack public?
      Anyone with the pack URL will be able to access it."
      @accepted="makePublic"
    />
    <ConfirmationDialog
      v-model="showPrivateDialog"
      title="Confirm Make Pack Private"
      text="Are you sure you want to make this pack private?
      Only assigned users will be able to access it."
      @accepted="makePrivate"
    />

    <v-card>
      <v-card-title
        class="d-flex flex-wrap align-center"
      >
        <div
          class="d-flex align-center me-auto"
        >
          <h1 class="me-5">
            {{ pack.name }}
          </h1>
          <PackStatus
            :status="pack.isArchived ? 'archived' : pack.status"
            class="me-2"
          />
          <PackStatus
            v-if="pack.isPublic"
            status="public"
            class="me-2"
          />
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

      <v-divider />

      <div
        class="d-flex flex-wrap ga-3 align-center justify-end mt-3 ms-3 me-3"
      >
        <v-btn
          v-if="!pack.isArchived && (pack.currentUserPermission >= PackPermission.EDIT || authStore.user?.isAdmin)"
          prepend-icon="mdi-pencil"
          text="Edit"
          :to="`/packs/${pack.slug}/edit`"
        />

        <div
          v-if="!pack.isArchived"
          class="d-flex align-center"
        >
          <v-btn
            v-if="pack.status === 'draft'"
            text="Publish"
            prepend-icon="mdi-earth"
            @click="showPublishDialog = true"
          />
          <v-btn
            v-else-if="pack.status === 'published'"
            text="Convert to Draft"
            color="warning"
            prepend-icon="mdi-file-edit"
            @click="showDraftDialog = true"
          />
        </div>

        <div class="d-flex align-center">
          <v-btn
            v-if="!pack.isArchived"
            text="Archive"
            color="error"
            prepend-icon="mdi-archive"
            @click="showArchiveDialog = true"
          />
          <v-btn
            v-else
            text="Unarchive"
            color="warning"
            prepend-icon="mdi-archive-refresh"
            @click="showUnArchiveDialog = true"
          />
        </div>

        <div
          v-if="!pack.isArchived"
          class="d-flex align-center"
        >
          <v-btn
            v-if="!pack.isPublic"
            text="Make Public"
            color="error"
            prepend-icon="mdi-earth"
            @click="showPublicDialog = true"
          />
          <v-btn
            v-else
            text="Make Private"
            color="warning"
            prepend-icon="mdi-earth-off"
            @click="showPrivateDialog = true"
          />
        </div>
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
      :pack-id="pack.id"
      :mods="pack.mods || []"
      :can-edit="pack.currentUserPermission >= PackPermission.EDIT && !pack.isArchived"
      @add-mod="onAddMod"
    />
  </div>
</template>
