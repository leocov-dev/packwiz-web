import "reflect-metadata"
import {afterEach, describe, expect, it, vi} from "vitest"

const get = vi.fn()

vi.mock("@/services/api.service", () => ({
  apiClient: {get},
}))

const originalNavigator = globalThis.navigator
const originalDocument = globalThis.document

afterEach(() => {
  vi.restoreAllMocks()
  Object.defineProperty(globalThis, 'navigator', {value: originalNavigator, configurable: true})
  Object.defineProperty(globalThis, 'document', {value: originalDocument, configurable: true})
})

describe("linkToClipboard", () => {
  it("falls back to the document copy command when the Clipboard API is unavailable", async () => {
    get.mockResolvedValue({data: {link: 'https://example.com/pack.toml'}})
    const textArea = {
      value: '',
      style: {},
      select: vi.fn(),
    }
    const appendChild = vi.fn()
    const removeChild = vi.fn()
    const execCommand = vi.fn().mockReturnValue(true)
    Object.defineProperty(globalThis, 'navigator', {value: {}, configurable: true})
    Object.defineProperty(globalThis, 'document', {
      value: {
        createElement: vi.fn().mockReturnValue(textArea),
        body: {appendChild, removeChild},
        execCommand,
      },
      configurable: true,
    })

    const {linkToClipboard} = await import("./packs.service.ts")
    await linkToClipboard(42)

    expect(get).toHaveBeenCalledWith('v1/packwiz/pack/42/link')
    expect(textArea.value).toBe('https://example.com/pack.toml')
    expect(textArea.select).toHaveBeenCalledOnce()
    expect(execCommand).toHaveBeenCalledWith('copy')
    expect(removeChild).toHaveBeenCalledWith(textArea)
  })
})