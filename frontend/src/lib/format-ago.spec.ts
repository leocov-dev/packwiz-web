import {describe, expect, it} from "vitest"
import {formatAgo} from "@/lib/format-ago.ts"

describe("formatAgo", () => {
  const now = Date.parse("2026-01-02T00:00:00Z")
  it("formats ranges", () => {
    expect(formatAgo("Seen", "2026-01-02T00:00:00Z", now)).toBe("Seen just now")
    expect(formatAgo("Seen", "2026-01-01T23:30:00Z", now)).toBe("Seen 30 min ago")
    expect(formatAgo("Seen", "2026-01-01T21:00:00Z", now)).toBe("Seen 3 h ago")
    expect(formatAgo("Seen", "2025-12-30T00:00:00Z", now)).toBe("Seen 3 d ago")
  })
  it("is empty for bad input", () => {
    expect(formatAgo("Seen", null, now)).toBe("")
    expect(formatAgo("Seen", "junk", now)).toBe("")
  })
})
