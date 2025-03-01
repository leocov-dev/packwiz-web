<route lang="yaml">
meta:
  layout: app
</route>


<script lang="ts" setup>
import {type Filters} from "@/components/FiltersMenu.vue";
import type {LocationQueryRaw} from "vue-router";


const route = useRoute();
const router = useRouter();

const queryData = ref<{ filters: Filters, search: string }>({
  filters: {
    'draft': true,
    'published': true,
  },
  search: '',
})

const updateFromUrl = () => {
  queryData.value.search = (route.query.search as string) || ''

  queryData.value.filters = (route.query.filters as string)?.split(',').reduce((map, key) => {
    map[key.trim()] = true
    return map
  }, {} as Filters) || {}
}
updateFromUrl()

watch(
  queryData,
  (newData) => {

    const query: LocationQueryRaw = {}

    if (!!newData.search) {
      query['search'] = newData.search
    }

    const filters = Object.keys(newData.filters).filter(k => newData.filters[k])

    if (filters.length > 0) {
      query['filters'] = filters.sort().join(',')
    }

    router.push({
      query: query
    })
  },
  {deep: true},
)

watch(
  () => route.query,
  () => {
    updateFromUrl()
  },
  {deep: true},
)
</script>

<template>
  <PackList
    v-model="queryData"
  />
</template>
