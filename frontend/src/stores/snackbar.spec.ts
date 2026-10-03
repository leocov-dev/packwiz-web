import {beforeEach, describe, expect, it} from "vitest"
import {createPinia, setActivePinia} from "pinia"
import {useSnackbarStore} from "@/stores/snackbar.ts"

describe("snackbar store", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it("shows a message without a link by default", () => {
    const store = useSnackbarStore()
    store.showSnackbar("Saved", "success")

    expect(store.show).toBe(true)
    expect(store.message).toBe("Saved")
    expect(store.link).toBeNull()
  })

  it("keeps the link it was given", () => {
    const store = useSnackbarStore()
    store.showSnackbar("Created", "success", 10000, {text: "View pack", to: "/packs/9"})

    expect(store.link).toEqual({text: "View pack", to: "/packs/9"})
    expect(store.timeout).toBe(10000)
  })

  it("does not carry a link over to the next message", () => {
    const store = useSnackbarStore()
    store.showSnackbar("Created", "success", 10000, {text: "View pack", to: "/packs/9"})
    store.showSnackbar("Something else", "info")

    expect(store.link).toBeNull()
  })

  it("clears the link when closed", () => {
    const store = useSnackbarStore()
    store.showSnackbar("Created", "success", 10000, {text: "View pack", to: "/packs/9"})
    store.closeSnackbar()

    expect(store.show).toBe(false)
    expect(store.link).toBeNull()
    expect(store.message).toBe("")
  })
})
