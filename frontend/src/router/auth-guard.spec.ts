import {describe, expect, it, vi, beforeEach} from "vitest"
import type {NavigationGuardNext, RouteLocationNormalizedGeneric, RouteLocationNormalizedLoaded} from "vue-router"

const checkAuth = vi.fn()
let isAuthenticated = false

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({
    checkAuth,
    get isAuthenticated() {
      return isAuthenticated
    },
  }),
}))

const {authGuard} = await import("./auth-guard.ts")

function makeTo(overrides: Partial<RouteLocationNormalizedGeneric> = {}): RouteLocationNormalizedGeneric {
  return {
    fullPath: "/packs",
    meta: {},
    ...overrides,
  } as RouteLocationNormalizedGeneric
}

const from = {} as RouteLocationNormalizedLoaded

beforeEach(() => {
  checkAuth.mockClear()
  isAuthenticated = false
})

describe("authGuard", () => {
  it("passes through to a noAuth route regardless of auth state", async () => {
    const next = vi.fn() as unknown as NavigationGuardNext
    await authGuard(makeTo({meta: {noAuth: true}}), from, next)

    expect(checkAuth).toHaveBeenCalledWith(false)
    expect(next).toHaveBeenCalledWith()
  })

  it("passes through to a protected route when authenticated", async () => {
    isAuthenticated = true
    const next = vi.fn() as unknown as NavigationGuardNext
    await authGuard(makeTo(), from, next)

    expect(next).toHaveBeenCalledWith()
  })

  it("redirects to login when not authenticated on a protected route", async () => {
    isAuthenticated = false
    const next = vi.fn() as unknown as NavigationGuardNext
    await authGuard(makeTo({fullPath: "/packs/1"}), from, next)

    expect(next).toHaveBeenCalledWith({
      path: "/auth/login",
      query: {redirect: "/packs/1"},
    })
  })
})
