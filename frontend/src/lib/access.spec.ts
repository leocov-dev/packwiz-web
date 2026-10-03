import {describe, expect, it} from "vitest";
import type {AccessDay} from "@/interfaces/access.ts";
import {formatDay, labelStep, seriesLabels, seriesTotal, seriesValues, withAlpha} from "@/lib/access.ts";

const series: AccessDay[] = [
  {date: "2026-10-01", success: 2, failure: 0},
  {date: "2026-10-02", success: 0, failure: 3},
  {date: "2026-10-03", success: 5, failure: 1},
]

describe("access series", () => {
  it("extracts values by kind", () => {
    expect(seriesValues(series, "success")).toEqual([2, 0, 5])
    expect(seriesValues(series, "failure")).toEqual([0, 3, 1])
  })

  it("totals by kind", () => {
    expect(seriesTotal(series, "success")).toBe(7)
    expect(seriesTotal(series, "failure")).toBe(4)
    expect(seriesTotal([], "success")).toBe(0)
  })

  it("formats days independent of time zone", () => {
    expect(formatDay("2026-10-03")).toBe("Oct 3")
    expect(formatDay("2026-01-01")).toBe("Jan 1")
    expect(formatDay("garbage")).toBe("garbage")
  })

  it("labels every nth day", () => {
    expect(seriesLabels(series, 2)).toEqual(["Oct 1", "", "Oct 3"])
  })

  it("picks a label step", () => {
    expect(labelStep(7)).toBe(1)
    expect(labelStep(30)).toBe(5)
    expect(labelStep(90)).toBe(13)
  })

  it("adds alpha to hex colors", () => {
    expect(withAlpha("#8b3eaa", 0.5)).toBe("rgba(139, 62, 170, 0.5)")
    expect(withAlpha("#fff", 0.2)).toBe("rgba(255, 255, 255, 0.2)")
    expect(withAlpha("rgb(1, 2, 3)", 0.2)).toBe("rgb(1, 2, 3)")
  })
})
