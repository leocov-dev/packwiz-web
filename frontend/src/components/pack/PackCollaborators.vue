<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts";
import {getPackCollaborators, listPackRoles} from "@/services/packs.service.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import CollaboratorCard from "@/components/pack/CollaboratorCard.vue";
import AddCollaboratorDialog from "@/components/pack/AddCollaboratorDialog.vue";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {Perm} from "@/lib/permissions.ts";

const {pack} = defineProps<{ pack: PackResponse }>()

const {can} = usePackPermissions(() => pack)
const canManage = computed(() => can(Perm.PackUsersManage))

const router = useRouter()

const search = ref<string>('')
const showAddDialog = ref(false)

const {
  data: collaboratorsResponse,
  reload,
} = buildDataLoader(async () => {
  return getPackCollaborators(pack.id)
})

const {data: rolesResponse} = buildDataLoader(async () => {
  return listPackRoles()
})

const roles = computed(() => rolesResponse.value?.roles || [])

const collaborators = computed(() => collaboratorsResponse.value?.users || [])

const backToPack = async () => {
  await router.push({path: `/packs/${pack.id}`})
}
</script>

<template>
  <div class="ma-6">
    <AddCollaboratorDialog
      v-if="canManage"
      v-model="showAddDialog"
      :roles="roles"
      :pack-id="pack.id"
      :existing-user-ids="collaborators.map(c => c.userId)"
      @added="reload"
    />

    <v-card>
      <v-card-title class="d-flex align-baseline">
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
        <h2>Collaborators</h2>
      </v-card-title>

      <v-data-iterator
        :items="collaborators"
        :search="search"
        items-per-page="0"
      >
        <template #header>
          <div class="d-flex flex-wrap align-center ga-3 pa-4 bg-surface-light">
            <v-text-field
              v-model="search"
              max-width="300"
              density="compact"
              placeholder="Search"
              prepend-inner-icon="mdi-magnify"
              variant="solo"
              clearable
              hide-details
            />
            <v-btn
              v-if="canManage"
              color="primary"
              variant="flat"
              prepend-icon="mdi-account-plus"
              text="Add Collaborator"
              @click="showAddDialog = true"
            />
          </div>
        </template>

        <template #default="{items}">
          <v-list>
            <v-list-item
              v-for="item in items"
              :key="item.raw.userId"
            >
              <CollaboratorCard
                :pack-id="pack.id"
                :collaborator="item.raw"
                :roles="roles"
                :can-manage="canManage"
                @changed="reload"
              />
            </v-list-item>
          </v-list>
        </template>

        <template #no-data>
          <p class="ma-6 text-disabled">
            No collaborators yet.
          </p>
        </template>
      </v-data-iterator>

      <v-card-actions class="ms-2 me-2 mb-2">
        <v-btn
          text="Back to Pack"
          @click="backToPack"
        />
      </v-card-actions>
    </v-card>
  </div>
</template>
