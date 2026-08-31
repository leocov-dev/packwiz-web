<script setup lang="ts">
import {fetchUsersPaginated} from "@/services/user.service.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {useRouter} from "vue-router";
import type {User} from "@/interfaces/user.ts";

const router = useRouter()

const nameSearch = ref('')
const emailSearch = ref('')
const userType = ref('')

const page = ref(1)
const itemsPerPage = ref(20)

const userTypeItems = [
  {title: 'All users', value: ''},
  {title: 'Admin', value: 'admin'},
  {title: 'User', value: 'user'},
]

const headers = [
  {title: 'Username', key: 'username'},
  {title: 'Full name', key: 'fullName'},
  {title: 'Email', key: 'email'},
  {title: 'Admin', key: 'isAdmin'},
  {title: 'Status', key: 'isActive'},
  {title: 'Created', key: 'createdAt'},
]

const {
  isLoading,
  data,
  reload,
} = buildDataLoader(async () => {
  return fetchUsersPaginated(page.value, itemsPerPage.value, nameSearch.value, emailSearch.value, userType.value)
})

const users = computed(() => data.value?.results || [])
const total = computed(() => data.value?.pagination.total || 0)

const onUpdateOptions = (options: { page: number, itemsPerPage: number }) => {
  page.value = options.page
  itemsPerPage.value = options.itemsPerPage
  reload()
}

watch([nameSearch, emailSearch, userType], () => {
  page.value = 1
  reload()
})

const onRowClick = (_: Event, {item}: { item: User }) => {
  router.push(`/admin/users/${item.id}`)
}
</script>

<template>
  <v-data-table-server
    :headers="headers"
    :items="users"
    :items-length="total"
    :items-per-page="itemsPerPage"
    :page="page"
    :loading="isLoading"
    class="pww-clickable-rows"
    @update:options="onUpdateOptions"
    @click:row="onRowClick"
  >
    <template #top>
      <v-toolbar
        class="ps-5 pe-5 pt-2 pb-2 d-flex flex-wrap"
        elevation="4"
      >
        <SearchBar
          v-model="nameSearch"
          max-width="260"
          class="me-3"
          density="comfortable"
        />
        <v-text-field
          v-model="emailSearch"
          max-width="260"
          class="me-3"
          density="comfortable"
          placeholder="Search by email"
          prepend-inner-icon="mdi-email"
          variant="solo"
          clearable
          hide-details
          @keyup.enter="reload"
          @click:clear="() => { emailSearch = ''; reload() }"
        />
        <v-select
          v-model="userType"
          :items="userTypeItems"
          max-width="180"
          class="me-auto"
          density="comfortable"
          variant="solo"
          hide-details
        />
        <v-btn
          text="New User"
          prepend-icon="mdi-plus"
          class="me-3"
          @click="router.push('/admin/users/new')"
        />
        <v-btn
          icon="mdi-refresh"
          @click="reload()"
        />
      </v-toolbar>
    </template>

    <template #item.isAdmin="{ item }">
      <v-chip
        :color="item.isAdmin ? 'primary' : undefined"
        :text="item.isAdmin ? 'Admin' : 'User'"
        size="small"
      />
    </template>

    <template #item.isActive="{ item }">
      <v-chip
        :color="item.isActive ? 'success' : 'error'"
        :text="item.isActive ? 'Active' : 'Deactivated'"
        size="small"
      />
    </template>

    <template #item.createdAt="{ item }">
      {{ new Date(item.createdAt).toLocaleDateString() }}
    </template>

    <template #loading>
      <v-skeleton-loader type="table-row@5" />
    </template>

    <template #no-data>
      <div class="d-flex justify-center ma-10">
        No results.
      </div>
    </template>
  </v-data-table-server>
</template>

<style scoped lang="sass">
.pww-clickable-rows :deep(tbody tr)
  cursor: pointer
</style>
