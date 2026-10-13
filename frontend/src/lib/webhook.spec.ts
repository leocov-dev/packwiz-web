import {describe, expect, it} from "vitest"
import {
  buildWebhookRequest, formFromResponse, lastAttemptSummary, validateWebhookForm, validateWebhookUrl,
  type WebhookFormData,
} from "@/lib/webhook.ts"

const form = (over: Partial<WebhookFormData> = {}): WebhookFormData => ({
  enabled: true,
  url: "https://cow.lan/api/servers/1/restart",
  method: "POST",
  headers: [{name: "Cookie", value: "sid=1", stored: false}],
  ...over,
})

describe("validateWebhookUrl", () => {
  it.each(["https://a.example/x", "http://192.168.1.5:3000/restart"])("accepts %s", (u) =>
    expect(validateWebhookUrl(u)).toBe(true))
  it.each(["", "cow.lan/x", "ftp://a/x", "https://u:p@a/x"])("rejects %s", (u) =>
    expect(validateWebhookUrl(u)).not.toBe(true))
})

describe("validateWebhookForm", () => {
  it("accepts a complete form", () => expect(validateWebhookForm(form())).toBeNull())

  it("accepts an empty disabled form", () =>
    expect(validateWebhookForm(form({enabled: false, url: "", headers: []}))).toBeNull())

  it("requires a url when enabled", () =>
    expect(validateWebhookForm(form({url: ""}))).not.toBeNull())

  it("allows a stored header with a blank value", () =>
    expect(validateWebhookForm(form({headers: [{name: "Cookie", value: "", stored: true}]}))).toBeNull())

  it("requires a value for a new header", () =>
    expect(validateWebhookForm(form({headers: [{name: "X-A", value: "", stored: false}]}))).toMatch(/needs a value/))

  it.each([
    [[{name: "bad name", value: "v", stored: false}]],
    [[{name: "Host", value: "v", stored: false}]],
    [[{name: "A", value: "1", stored: false}, {name: "a", value: "2", stored: false}]],
    [[{name: "A", value: "x\ny", stored: false}]],
  ])("rejects bad headers %#", (headers) =>
    expect(validateWebhookForm(form({headers}))).not.toBeNull())
})

describe("buildWebhookRequest", () => {
  it("trims, drops blank rows and keeps blank values of stored headers", () => {
    const req = buildWebhookRequest(form({
      url: "  https://a.example/x ",
      headers: [
        {name: " Cookie ", value: "", stored: true},
        {name: "", value: "", stored: false},
      ],
    }))
    expect(req).toEqual({
      enabled: true,
      url: "https://a.example/x",
      method: "POST",
      headers: [{name: "Cookie", value: ""}],
    })
  })
})

describe("formFromResponse", () => {
  it("never carries values and marks headers as stored", () => {
    const f = formFromResponse({
      enabled: true, url: "https://a/x", method: "PUT", headers: [{name: "Cookie"}], broken: false,
      lastStatusCode: 0, lastError: "", lastAttemptAt: null, lastSuccessAt: null,
    })
    expect(f.headers).toEqual([{name: "Cookie", value: "", stored: true}])
    expect(f.method).toBe("PUT")
  })
})

describe("lastAttemptSummary", () => {
  it("covers never, failed and succeeded", () => {
    expect(lastAttemptSummary({lastAttemptAt: null, lastError: "", lastStatusCode: 0})).toBe("Not called yet.")
    const at = "2026-10-12T10:00:00Z"
    expect(lastAttemptSummary({lastAttemptAt: at, lastError: "HTTP 500", lastStatusCode: 500})).toMatch(/^Failed .*HTTP 500$/)
    expect(lastAttemptSummary({lastAttemptAt: at, lastError: "", lastStatusCode: 200})).toMatch(/\(HTTP 200\)\.$/)
  })
})
