<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts";
import PackActions from "@/components/pack/PackActions.vue";
import ModsList from "@/components/mods/ModsList.vue";
import {toTitleCase} from "@/services/utils.ts";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {Perm} from "@/lib/permissions.ts";
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
import {
  countUpdatable,
  formatCheckedAgo,
  shouldForceCheck,
  updateAllLabel,
  updatesAvailableText
} from "@/lib/update-checks.ts";
import PackAccessSparkline from "@/components/pack/PackAccessSparkline.vue";
import {applyPinOverrides} from "@/lib/mod-filters.ts";

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
const {can} = usePackPermissions(() => pack)

const canManageMenu = computed(() => [
  Perm.PackUsersView, Perm.PackSnapshotView, Perm.PackView, Perm.PackMigrate, Perm.PackRehash, Perm.PackPublish, Perm.PackVisibility,
  Perm.PackArchive,
].some(can))
const hasToolbar = computed(() => canManageMenu.value || [
  Perm.PackInfoEdit, Perm.PackModAdd, Perm.PackUpdatesCheck, Perm.PackModUpdate,
].some(can))

const updateChecks = useUpdateChecks(() => pack.id)
// Optimistic pin changes reported by ModsList (not yet in pack.mods); applied so
// "Update All (N)" matches the badges without a reload.
const pinOverrides = ref<ReadonlyMap<number, boolean>>(new Map())
watch(() => pack.mods, () => {
  pinOverrides.value = new Map()
})
const effectiveMods = computed(() => applyPinOverrides(pack.mods ?? [], pinOverrides.value))
const updatableCount = computed(() => countUpdatable(effectiveMods.value, updateChecks.results.value))
const hasCompletedCheck = computed(() => updateChecks.status.value === "done")
const updateAllText = computed(() => hasCompletedCheck.value ? updateAllLabel(updatableCount.value) : "Update All")
const checkedText = computed(() => {
  if (updateChecks.isChecking.value) return "Checking for updates..."
  if (updateChecks.error.value) return updateChecks.error.value
  if (!hasCompletedCheck.value) return ""
  return `${updatesAvailableText(updatableCount.value)} · ${formatCheckedAgo(updateChecks.checkedAt.value)}`
})
// Force only once the cached result expired; within the TTL the server serves the
// cache, and it enforces a minimum interval between runs regardless of force.
const onCheckUpdates = () =>
  updateChecks.start(shouldForceCheck(updateChecks.status.value, updateChecks.checkedAt.value))

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
      v-if="can(Perm.PackMigrate)"
      v-model="showMigrateDialog"
      :pack="pack"
      @migrated="emit('reload')"
    />
    <RehashDialog
      v-if="can(Perm.PackRehash)"
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
          <v-tooltip
            text="Back to packs"
            location="bottom"
          >
            <template #activator="{ props }">
              <v-btn
                v-bind="props"
                icon="mdi-arrow-left"
                variant="text"
                class="me-2"
                aria-label="Back to packs"
                to="/packs"
              />
            </template>
          </v-tooltip>
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
          <router-link
            v-if="can(Perm.PackView)"
            :to="`/packs/${pack.id}/access`"
            class="d-none d-md-flex ms-3"
            aria-label="View pack access metrics"
          >
            <PackAccessSparkline :pack-id="pack.id" />
          </router-link>
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
        v-if="hasToolbar"
        class="d-flex flex-wrap ga-3 align-center justify-end mt-3 ms-3 me-3"
      >
        <v-btn
          v-if="can(Perm.PackInfoEdit)"
          prepend-icon="mdi-pencil"
          text="Edit"
          :to="`/packs/${pack.id}/edit`"
        />
        <v-btn
          v-if="can(Perm.PackModAdd)"
          prepend-icon="mdi-plus"
          text="Add Mod"
          color="primary"
          variant="flat"
          @click="onAddMod"
        />
        <v-btn
          v-if="can(Perm.PackUpdatesCheck)"
          prepend-icon="mdi-cloud-search-outline"
          text="Check for updates"
          :loading="updateChecks.isChecking.value"
          :disabled="updateChecks.isChecking.value || updateChecks.coolingDown.value || updateAllLoading"
          @click="onCheckUpdates"
        />
        <v-btn
          v-if="can(Perm.PackModUpdate)"
          prepend-icon="mdi-update"
          :text="updateAllText"
          :loading="updateAllLoading"
          :disabled="updateAllLoading"
          @click="showUpdateAllDialog = true"
        />

        <v-menu v-if="canManageMenu">
          <template #activator="{ props }">
            <v-btn
              v-bind="props"
              prepend-icon="mdi-cog"
              append-icon="mdi-menu-down"
              text="Manage"
            />
          </template>

          <v-list>
            <v-list-item
              v-if="can(Perm.PackUsersView)"
              prepend-icon="mdi-account-multiple"
              title="Collaborators"
              :to="`/packs/${pack.id}/collaborators`"
            />
            <v-list-item
              v-if="can(Perm.PackSnapshotView)"
              prepend-icon="mdi-history"
              title="History"
              :to="`/packs/${pack.id}/snapshots`"
            />
            <v-list-item
              v-if="can(Perm.PackView)"
              prepend-icon="mdi-chart-bar"
              title="Access"
              :to="`/packs/${pack.id}/access`"
            />
            <v-list-item
              v-if="can(Perm.PackMigrate)"
              prepend-icon="mdi-arrow-up-bold-hexagon-outline"
              title="Migrate"
              @click="showMigrateDialog = true"
            />
            <v-list-item
              v-if="can(Perm.PackRehash)"
              prepend-icon="mdi-pound-box-outline"
              title="Rehash"
              @click="showRehashDialog = true"
            />
            <v-list-item
              v-if="can(Perm.PackPublish) && pack.status === 'draft'"
              prepend-icon="mdi-earth"
              title="Publish"
              @click="showPublishDialog = true"
            />
            <v-list-item
              v-else-if="can(Perm.PackPublish) && pack.status === 'published'"
              prepend-icon="mdi-file-edit"
              title="Convert to Draft"
              @click="showDraftDialog = true"
            />
            <v-list-item
              v-if="can(Perm.PackVisibility) && !pack.isPublic"
              prepend-icon="mdi-earth"
              title="Make Public"
              @click="showPublicDialog = true"
            />
            <v-list-item
              v-else-if="can(Perm.PackVisibility)"
              prepend-icon="mdi-earth-off"
              title="Make Private"
              @click="showPrivateDialog = true"
            />
            <v-divider class="my-1" />
            <v-list-item
              v-if="can(Perm.PackArchive) && !pack.isArchived"
              prepend-icon="mdi-archive"
              title="Archive"
              base-color="error"
              @click="showArchiveDialog = true"
            />
            <v-list-item
              v-if="can(Perm.PackArchive) && pack.isArchived"
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

    <v-sheet
      v-if="checkedText"
      :color="updateChecks.error.value ? 'warning' : 'primary'"
      class="text-body-2 my-1 px-4 py-2"
      role="status"
    >
      {{ checkedText }}
    </v-sheet>

    <ModsList
      :pack-id="pack.id"
      :mods="pack.mods || []"
      :permissions="pack.permissions"
      :update-checks="updateChecks.results.value"
      @add-mod="onAddMod"
      @pin-overrides="pinOverrides = $event"
      @reload="$emit('reload')"
    />
  </div>
</template>
