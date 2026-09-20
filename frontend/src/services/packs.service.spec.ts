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
  it.each([
    ["unavailable", false],
    ["rejected", true],
  ])("falls back to the document copy command when the Clipboard API is %s", async (_state, rejects) => {
    get.mockResolvedValue({data: {link: 'https://example.com/pack.toml'}})
    const textArea = {
      value: '',
      style: {},
      select: vi.fn(),
    }
    const appendChild = vi.fn()
    const removeChild = vi.fn()
    const execCommand = vi.fn().mockReturnValue(true)
    const clipboard = !rejects
      ? undefined
      : {writeText: vi.fn().mockRejectedValue(new Error("Clipboard permission denied"))}
    Object.defineProperty(globalThis, 'navigator', {value: {clipboard}, configurable: true})
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
