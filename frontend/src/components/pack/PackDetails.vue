<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts";
import PackActions from "@/components/pack/PackActions.vue";
import ModsList from "@/components/mods/ModsList.vue";
import {toTitleCase} from "@/services/utils.ts";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {
  archivePack,
  convertPackToDraft,
  makePackPrivate,
  makePackPublic,
  publishPack,
  unArchivePack,
  updateAllMods
} from "@/services/packs.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import PackMigrateDialog from "@/components/pack/PackMigrateDialog.vue";
import RehashDialog from "@/components/pack/RehashDialog.vue";
import UpdateAllResultDialog from "@/components/pack/UpdateAllResultDialog.vue";
import {summarizeUpdateAll, updateAllOutcome} from "@/lib/update-summary.ts";
import type {UpdateAllResponse} from "@/interfaces/pack.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {useUpdateChecks} from "@/composables/useUpdateChecks.ts";
import {countUpdatable, formatCheckedAgo, updateAllLabel, updatesAvailableText} from "@/lib/update-checks.ts";

const {pack} = defineProps<{ pack: PackResponse }>()

const emit = defineEmits(['reload'])

const showPublishDialog = ref(false)
const showDraftDialog = ref(false)
const showArchiveDialog = ref(false)
const showUnArchiveDialog = ref(false)
const showPublicDialog = ref(false)
const showPrivateDialog = ref(false)
const showUpdateAllDialog = ref(false)
const showMigrateDialog = ref(false)
const showRehashDialog = ref(false)

const updateAllLoading = ref(false)
const showUpdateAllResult = ref(false)
const updateAllResult = ref<UpdateAllResponse | null>(null)

const snackbar = useSnackbarStore()

const router = useRouter()
const {canEdit, hasEditAccess} = usePackPermissions(() => pack)

const updateChecks = useUpdateChecks(() => pack.id)
const updatableCount = computed(() => countUpdatable(pack.mods ?? [], updateChecks.results.value))
const hasCompletedCheck = computed(() => updateChecks.status.value === "done")
const updateAllText = computed(() => hasCompletedCheck.value ? updateAllLabel(updatableCount.value) : "Update All")
const checkedText = computed(() => {
  if (updateChecks.isChecking.value) return "Checking for updates..."
  if (updateChecks.error.value) return updateChecks.error.value
  if (!hasCompletedCheck.value) return ""
  return `${updatesAvailableText(updatableCount.value)} · ${formatCheckedAgo(updateChecks.checkedAt.value)}`
})
const onCheckUpdates = () => updateChecks.start(hasCompletedCheck.value)

const onAddMod = () => {
  router.push({path: `/packs/${pack.id}/add-mod`})
}

const chipList = computed<string[]>(() => {
  const chips: string[] = []
  if (pack.slug) chips.push(`Slug: ${pack.slug}`)
  if (pack.version) chips.push(`Version: ${pack.version}`)
  if (pack.mcVersion) chips.push(`Minecraft: ${pack.mcVersion}`)
  if (pack.loader && pack.loaderVersion) chips.push(`${toTitleCase(pack.loader)}: ${pack.loaderVersion}`)
  if (pack.packFormat) chips.push(`Format: ${pack.packFormat}`)

  const gameVersions = pack.acceptableGameVersions ?? []
  const repeatsMinecraft = gameVersions.length === 0
    || (gameVersions.length === 1 && gameVersions[0] === pack.mcVersion)
  if (!repeatsMinecraft) chips.push(`Game versions: ${gameVersions.join(', ')}`)
  return chips
})


const convertToDraft = async () => {
  await convertPackToDraft(pack.id)
  emit('reload')
}

const publish = async () => {
  await publishPack(pack.id)
  emit('reload')
}

const archive = async () => {
  await archivePack(pack.id)
  emit('reload')
}

const unArchive = async () => {
  await unArchivePack(pack.id)
  emit('reload')
}

const makePublic = async () => {
  await makePackPublic(pack.id)
  emit('reload')
}

const makePrivate = async () => {
  await makePackPrivate(pack.id)
  emit('reload')
}

const updateAll = async () => {
  updateAllLoading.value = true
  try {
    const result = await updateAllMods(pack.id)
    if (updateAllOutcome(result) === "up-to-date") {
      snackbar.showSnackbar(summarizeUpdateAll(result), "success")
    } else {
      updateAllResult.value = result
      showUpdateAllResult.value = true
    }
  } catch (e) {
    snackbar.showSnackbar(apiErrorMessage(e, "Failed to update mods"), "error")
  } finally {
    updateAllLoading.value = false
    // the mods changed, so stored check results are stale (the server clears them too)
    updateChecks.reset()
    // always reload: a failed/timed-out run may still have applied some updates
    emit('reload')
  }
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
    <ConfirmationDialog
      v-model="showUpdateAllDialog"
      title="Confirm Update All Mods"
      text="Are you sure you want to update all mods to their latest versions?
      Pinned mods will be skipped."
      @accepted="updateAll"
    />
    <UpdateAllResultDialog
      v-if="updateAllResult"
      v-model="showUpdateAllResult"
      :result="updateAllResult"
    />
    <PackMigrateDialog
      v-if="canEdit"
      v-model="showMigrateDialog"
      :pack="pack"
      @migrated="emit('reload')"
    />
    <RehashDialog
      v-if="canEdit"
      v-model="showRehashDialog"
      :pack="pack"
      @rehashed="emit('reload')"
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
        <v-tooltip
          text="Reload pack"
          location="bottom"
        >
          <template #activator="{ props }">
            <v-btn
              v-bind="props"
              icon="mdi-refresh"
              variant="text"
              class="text-medium-emphasis"
              aria-label="Reload pack"
              @click="$emit('reload')"
            />
          </template>
        </v-tooltip>
      </v-card-title>

      <v-divider />

      <div
        v-if="hasEditAccess"
        class="d-flex flex-wrap ga-3 align-center justify-end mt-3 ms-3 me-3"
      >
        <template v-if="canEdit">
          <v-btn
            prepend-icon="mdi-pencil"
            text="Edit"
            :to="`/packs/${pack.id}/edit`"
          />
          <v-btn
            prepend-icon="mdi-plus"
            text="Add Mod"
            color="primary"
            variant="flat"
            @click="onAddMod"
          />
          <span
            v-if="checkedText"
            class="text-body-2 me-auto"
            :class="updateChecks.error.value ? 'text-warning' : 'text-medium-emphasis'"
            role="status"
          >
            {{ checkedText }}
          </span>
          <v-btn
            prepend-icon="mdi-cloud-search-outline"
            text="Check for updates"
            :loading="updateChecks.isChecking.value"
            :disabled="updateChecks.isChecking.value || updateAllLoading"
            @click="onCheckUpdates"
          />
          <v-btn
            prepend-icon="mdi-update"
            :text="updateAllText"
            :loading="updateAllLoading"
            :disabled="updateAllLoading"
            @click="showUpdateAllDialog = true"
          />
        </template>

        <v-menu>
          <template #activator="{ props }">
            <v-btn
              v-bind="props"
              prepend-icon="mdi-cog"
              append-icon="mdi-menu-down"
              text="Manage"
            />
          </template>

          <v-list>
            <template v-if="canEdit">
              <v-list-item
                prepend-icon="mdi-account-multiple"
                title="Collaborators"
                :to="`/packs/${pack.id}/collaborators`"
              />
              <v-list-item
                prepend-icon="mdi-arrow-up-bold-hexagon-outline"
                title="Migrate"
                @click="showMigrateDialog = true"
              />
              <v-list-item
                prepend-icon="mdi-pound-box-outline"
                title="Rehash"
                @click="showRehashDialog = true"
              />
              <v-list-item
                v-if="pack.status === 'draft'"
                prepend-icon="mdi-earth"
                title="Publish"
                @click="showPublishDialog = true"
              />
              <v-list-item
                v-else-if="pack.status === 'published'"
                prepend-icon="mdi-file-edit"
                title="Convert to Draft"
                @click="showDraftDialog = true"
              />
              <v-list-item
                v-if="!pack.isPublic"
                prepend-icon="mdi-earth"
                title="Make Public"
                @click="showPublicDialog = true"
              />
              <v-list-item
                v-else
                prepend-icon="mdi-earth-off"
                title="Make Private"
                @click="showPrivateDialog = true"
              />
              <v-divider class="my-1" />
              <v-list-item
                prepend-icon="mdi-archive"
                title="Archive"
                base-color="error"
                @click="showArchiveDialog = true"
              />
            </template>
            <v-list-item
              v-else
              prepend-icon="mdi-archive-refresh"
              title="Unarchive"
              base-color="warning"
              @click="showUnArchiveDialog = true"
            />
          </v-list>
        </v-menu>
      </div>

      <v-card-text class="ms-2 me-2 mb-2">
        <div
          v-if="chipList.length > 0"
          class="d-flex flex-wrap align-center"
        >
          <v-chip
            v-for="chip in chipList"
            :key="chip"
            class="me-2 mb-2"
            variant="tonal"
          >
            {{ chip }}
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
      :can-edit="canEdit"
      :update-checks="updateChecks.results.value"
      @add-mod="onAddMod"
      @reload="$emit('reload')"
    />
  </div>
</template>
