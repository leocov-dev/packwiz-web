import type {LocationQuery, LocationQueryRaw} from "vue-router"
import type {Filters} from "@/components/FiltersMenu.vue"

export interface PackFilterState {
  filters: Filters
  search: string
}

const DEFAULT_FILTERS: Filters = {
  'draft': true,
  'published': true,
}

export function queryToFilters(query: LocationQuery): PackFilterState {
  const search = (query.search as string) || ''

  const filters = (query.filters as string)?.split(',').reduce((map, key) => {
    map[key.trim()] = true
    return map
  }, {} as Filters) || {...DEFAULT_FILTERS}

  return {filters, search}
}

export function filtersToQuery(data: PackFilterState): LocationQueryRaw {
  const query: LocationQueryRaw = {}

  if (!!data.search) {
    query['search'] = data.search
  }

  const active = Object.keys(data.filters).filter(k => data.filters[k])

  if (active.length > 0) {
    query['filters'] = active.sort().join(',')
  }

  return query
}
