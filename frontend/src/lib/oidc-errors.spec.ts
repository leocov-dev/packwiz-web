import {describe, expect, it} from "vitest"
import {oidcErrorMessage} from "@/lib/oidc-errors.ts"

describe("oidcErrorMessage", () => {
  it("returns null when there is no error", () => {
    for (const v of [undefined, null, "", [], 3]) expect(oidcErrorMessage(v)).toBeNull()
  })

  it("maps known codes", () => {
    expect(oidcErrorMessage("oidc_not_permitted")).toContain("not permitted")
    expect(oidcErrorMessage("oidc_unavailable")).toContain("not available")
    expect(oidcErrorMessage("oidc_already_linked")).toContain("already linked")
  })

  it("uses the first value of a repeated parameter", () => {
    expect(oidcErrorMessage(["oidc_failed", "x"])).toContain("failed")
  })

  it("never echoes unknown codes", () => {
    const msg = oidcErrorMessage("<script>alert(1)</script>")
    expect(msg).not.toContain("script")
    expect(msg).toContain("Sign-in failed")
  })

  it("does not treat prototype keys as codes", () => {
    expect(oidcErrorMessage("constructor")).toContain("Sign-in failed")
  })
})
