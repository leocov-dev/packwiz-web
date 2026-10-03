<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts";
import type {PackSnapshot, PackSnapshotDetailResponse, SnapshotAgainst} from "@/interfaces/snapshot.ts";
import SnapshotDiff from "@/components/pack/SnapshotDiff.vue";
import CloneSnapshotDialog from "@/components/pack/CloneSnapshotDialog.vue";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {usePermissions} from "@/composables/usePermissions.ts";
import {Perm} from "@/lib/permissions.ts";
import {
  canRevertSnapshot,
  revertConfirmText,
  revertLabel,
  snapshotReasonIcon,
  snapshotReasonLabel,
  snapshotSubject,
  summarizeSnapshot,
} from "@/lib/snapshots.ts";
import {fetchPackSnapshot, revertToSnapshot} from "@/services/snapshots.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const open = defineModel<boolean>({required: true})

const {pack, snapshot} = defineProps<{
  pack: PackResponse
  snapshot: PackSnapshot | null
}>()

const emit = defineEmits<{ reverted: [] }>()

const {can} = usePackPermissions(() => pack)
const {can: canGlobal} = usePermissions()
const snackbar = useSnackbarStore()

const against = ref<SnapshotAgainst>('parent')
const detail = ref<PackSnapshotDetailResponse | null>(null)
const isLoading = ref(false)
const error = ref('')
// responses of a superseded request are dropped
let loadGeneration = 0

const canRevert = computed(() =>
  snapshot !== null && canRevertSnapshot(snapshot, can(Perm.PackSnapshotRevert)),
)
const canClone = computed(() => can(Perm.PackSnapshotView) && canGlobal(Perm.PackCreate))

const showRevertConfirm = ref(false)
const showClone = ref(false)
const isReverting = ref(false)

const load = async () => {
  if (!snapshot) {
    return
  }

  const generation = ++loadGeneration
  isLoading.value = true
  error.value = ''
  try {
    const result = await fetchPackSnapshot(pack.id, snapshot.id, against.value)
    if (generation === loadGeneration) {
      detail.value = result
    }
  } catch (e) {
    if (generation === loadGeneration) {
      detail.value = null
      error.value = apiErrorMessage(e, 'Failed to load the snapshot.')
    }
  } finally {
    if (generation === loadGeneration) {
      isLoading.value = false
    }
  }
}

watch(() => [open.value, snapshot?.id] as const, ([isOpen]) => {
  if (!isOpen || !snapshot) {
    return
  }
  against.value = 'parent'
  detail.value = null
  load()
}, {immediate: true})

watch(against, load)

const doRevert = async () => {
  if (!snapshot) {
    return
  }

  isReverting.value = true
  error.value = ''
  try {
    const result = await revertToSnapshot(pack.id, snapshot.id)
    snackbar.showSnackbar(
      result.changed ? `Reverted to snapshot #${snapshot.seq}` : 'The pack already matches this snapshot',
      'success',
      4000,
    )
    open.value = false
    emit('reverted')
  } catch (e) {
    error.value = apiErrorMessage(e, 'Failed to revert the pack.')
  } finally {
    isReverting.value = false
  }
}
</script>

<template>
  <v-dialog
    v-model="open"
    max-width="900"
    scrollable
  >
    <v-card
      v-if="snapshot"
      class="pa-3"
    >
      <v-card-title class="d-flex align-center flex-wrap ga-3">
        <span>Snapshot #{{ snapshot.seq }}</span>
        <v-chip
          size="small"
          label
          variant="tonal"
          :prepend-icon="snapshotReasonIcon(snapshot.reason)"
          :text="snapshotReasonLabel(snapshot.reason)"
        />
        <v-chip
          v-if="snapshot.isHead"
          size="small"
          label
          variant="tonal"
          color="primary"
          text="Current"
        />
        <v-chip
          v-if="snapshot.isAbandoned"
          size="small"
          label
          variant="tonal"
          text="Abandoned"
        />
      </v-card-title>
      <v-card-subtitle class="text-wrap">
        {{ new Date(snapshot.createdAt).toLocaleString() }}
        · {{ snapshot.createdByUsername || 'unknown user' }}
        <template v-if="snapshotSubject(snapshot.detail)">
          · {{ snapshotSubject(snapshot.detail) }}
        </template>
        · {{ summarizeSnapshot(snapshot.summary) }}
      </v-card-subtitle>

      <v-card-text>
        <v-alert
          v-if="snapshot.isAbandoned"
          class="mb-4"
          type="warning"
          variant="tonal"
          text="A revert to an earlier snapshot abandoned this one. It can be viewed and cloned, but not restored."
        />

        <v-btn-toggle
          v-model="against"
          class="mb-4"
          mandatory
          divided
          variant="tonal"
        >
          <v-btn
            value="parent"
            text="Changes in this snapshot"
          />
          <v-btn
            value="current"
            text="Changes since this snapshot"
          />
        </v-btn-toggle>

        <v-skeleton-loader
          v-if="isLoading"
          type="list-item-two-line@4"
        />
        <v-alert
          v-else-if="error"
          type="error"
          variant="tonal"
          :text="error"
        />
        <SnapshotDiff
          v-else-if="detail"
          :diff="detail.diff"
        />
      </v-card-text>

      <v-card-actions>
        <v-btn
          text="Close"
          variant="text"
          @click="open = false"
        />
        <v-spacer />
        <v-btn
          v-if="canClone"
          text="Clone as new pack"
          variant="tonal"
          prepend-icon="mdi-content-copy"
          @click="showClone = true"
        />
        <v-btn
          v-if="canRevert"
          :text="revertLabel(snapshot)"
          variant="tonal"
          color="error"
          prepend-icon="mdi-history"
          :loading="isReverting"
          @click="showRevertConfirm = true"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <template v-if="snapshot">
    <ConfirmationDialog
      v-model="showRevertConfirm"
      title="Revert pack?"
      :text="revertConfirmText(snapshot)"
      accept-text="Revert"
      @accepted="doRevert"
    />
    <CloneSnapshotDialog
      v-model="showClone"
      :pack="pack"
      :snapshot="snapshot"
    />
  </template>
</template>
