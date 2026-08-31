<route lang="yaml">
meta:
  layout: app
</route>


<script lang="ts" setup>
import {type Filters} from "@/components/FiltersMenu.vue";
import {queryToFilters, filtersToQuery} from "@/lib/pack-filters.ts";


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
  queryData.value = queryToFilters(route.query)
}
updateFromUrl()

watch(
  queryData,
  (newData) => {
    router.push({
      query: filtersToQuery(newData)
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
