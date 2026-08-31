import {describe, expect, it} from "vitest"
import {reduceFilters} from "./FiltersMenu.vue"

describe("reduceFilters", () => {
  it("returns an empty array for an empty object", () => {
    expect(reduceFilters({})).toEqual([])
  })

  it("returns all keys when all are true", () => {
    expect(reduceFilters({draft: true, published: true})).toEqual(['draft', 'published'])
  })

  it("returns only the true keys when mixed", () => {
    expect(reduceFilters({draft: true, published: false, archived: true})).toEqual(['draft', 'archived'])
  })

  it("returns an empty array when all are false", () => {
    expect(reduceFilters({draft: false, published: false})).toEqual([])
  })
})
