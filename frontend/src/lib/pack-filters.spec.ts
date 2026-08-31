import {describe, expect, it} from "vitest"
import {queryToFilters, filtersToQuery} from "./pack-filters.ts"
import type {LocationQuery} from "vue-router"

describe("queryToFilters", () => {
  it("defaults to draft+published and empty search for an empty query", () => {
    expect(queryToFilters({} as LocationQuery)).toEqual({
      filters: {draft: true, published: true},
      search: '',
    })
  })

  it("parses search only", () => {
    expect(queryToFilters({search: 'sodium'} as unknown as LocationQuery)).toEqual({
      filters: {draft: true, published: true},
      search: 'sodium',
    })
  })

  it("parses a single filter", () => {
    expect(queryToFilters({filters: 'draft'} as unknown as LocationQuery)).toEqual({
      filters: {draft: true},
      search: '',
    })
  })

  it("parses multiple comma-joined filters", () => {
    expect(queryToFilters({filters: 'draft,archived'} as unknown as LocationQuery)).toEqual({
      filters: {draft: true, archived: true},
      search: '',
    })
  })

  it("ignores unknown extra query params", () => {
    expect(queryToFilters({foo: 'bar'} as unknown as LocationQuery)).toEqual({
      filters: {draft: true, published: true},
      search: '',
    })
  })
})

describe("filtersToQuery", () => {
  it("omits search when empty and serializes default filters", () => {
    expect(filtersToQuery({filters: {draft: true, published: true}, search: ''}))
      .toEqual({filters: 'draft,published'})
  })

  it("includes search when set", () => {
    expect(filtersToQuery({filters: {draft: true, published: true}, search: 'sodium'}))
      .toEqual({filters: 'draft,published', search: 'sodium'})
  })

  it("omits the filters key entirely when all filters are false", () => {
    // Known asymmetry, preserved as-is: round-tripping an all-false filter
    // state through the URL does not survive, since queryToFilters falls
    // back to the default {draft:true, published:true} when the `filters`
    // query param is absent. Not a bug introduced here — existing behavior.
    expect(filtersToQuery({filters: {draft: false, published: false}, search: ''}))
      .toEqual({})
  })
})
