import "reflect-metadata"
import {beforeEach, describe, expect, it, vi} from "vitest"

const {get, put, post} = vi.hoisted(() => ({get: vi.fn(), put: vi.fn(), post: vi.fn()}))
vi.mock("@/services/api.service", () => ({apiClient: {get, put, post}}))

import {getPackWebhook, savePackWebhook, testPackWebhook} from "@/services/webhook.service.ts"

describe("webhook.service", () => {
  beforeEach(() => vi.resetAllMocks())

  it("gets the webhook", async () => {
    get.mockResolvedValue({data: {url: "u"}})
    expect(await getPackWebhook(3)).toEqual({url: "u"})
    expect(get).toHaveBeenCalledWith("v1/packwiz/pack/3/webhook")
  })

  it("saves the webhook", async () => {
    put.mockResolvedValue({data: {url: "u"}})
    const req = {enabled: true, url: "u", method: "POST" as const, headers: []}
    await savePackWebhook(3, req)
    expect(put).toHaveBeenCalledWith("v1/packwiz/pack/3/webhook", req)
  })

  it("tests the webhook", async () => {
    post.mockResolvedValue({data: {ok: true, statusCode: 200, error: ""}})
    expect((await testPackWebhook(3)).ok).toBe(true)
    expect(post).toHaveBeenCalledWith("v1/packwiz/pack/3/webhook/test")
  })
})
