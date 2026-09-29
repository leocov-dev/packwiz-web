import {describe, expect, it} from "vitest"
import {AxiosError} from "axios"
import {apiErrorMessage} from "@/services/utils.ts"

describe("apiErrorMessage", () => {
  it("uses the api error field", () => {
    const e = new AxiosError("x", "400", undefined, undefined, {data: {error: "nope"}, status: 400, statusText: "", headers: {}, config: {} as never})
    expect(apiErrorMessage(e, "fallback")).toBe("nope")
  })
  it("falls back for axios errors without body", () => {
    expect(apiErrorMessage(new AxiosError("x"), "fallback")).toBe("fallback")
  })
  it("stringifies other errors", () => {
    expect(apiErrorMessage(new Error("bad"), "fallback")).toBe("bad")
    expect(apiErrorMessage("oops", "fallback")).toBe("oops")
  })
})
