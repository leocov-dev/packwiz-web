<script setup lang="ts">
import type {SystemRole, User} from "@/interfaces/user.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {usePermissions} from "@/composables/usePermissions.ts";
import {Perm} from "@/lib/permissions.ts";
import {listSystemRoles, setUserRoles} from "@/services/user.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const {user} = defineProps<{ user: User }>()

const emit = defineEmits<{ changed: [] }>()

const snackbarStore = useSnackbarStore()
const {can} = usePermissions()

// the superuser's roles are fixed by the backend
const canAssign = computed(() => can(Perm.UserRolesAssign) && !user.isSuperuser)

const {data: roleData} = buildDataLoader<SystemRole[]>(() => listSystemRoles())

const roleItems = computed(() =>
  (roleData.value ?? [])
    .filter(r => r.assignable)
    .map(r => ({title: r.name, subtitle: r.description, value: r.id})),
)

const selected = ref<number[]>((user.roles ?? []).map(r => r.id))
const saving = ref(false)

const hasChanged = computed(() => {
  const current = new Set((user.roles ?? []).map(r => r.id))
  return current.size !== selected.value.length || selected.value.some(id => !current.has(id))
})

const save = async () => {
  saving.value = true
  try {
    await setUserRoles(user.id, selected.value)
    snackbarStore.showSnackbar('Roles updated.', 'success')
    emit('changed')
  } catch (e) {
    snackbarStore.showSnackbar(apiErrorMessage(e, 'Failed to update roles.'), 'error')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <v-divider />

  <v-card-subtitle class="mt-4">
    Roles
  </v-card-subtitle>

  <div class="ma-4">
    <div
      v-if="!canAssign"
      class="d-flex flex-wrap ga-2"
    >
      <v-chip
        v-if="user.isSuperuser"
        color="warning"
        text="Superuser"
        label
        variant="tonal"
      />
      <v-chip
        v-for="role in user.roles"
        :key="role.id"
        :text="role.name"
        label
        variant="tonal"
      />
    </div>

    <template v-else>
      <v-select
        v-model="selected"
        :items="roleItems"
        label="System roles"
        multiple
        chips
        closable-chips
        hide-details
      />
      <div
        v-if="hasChanged"
        class="d-flex justify-end mt-3"
      >
        <v-btn
          text="Save roles"
          :loading="saving"
          :disabled="saving"
          @click="save"
        />
      </div>
    </template>
  </div>
</template>
