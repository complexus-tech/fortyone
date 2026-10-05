import { listPresets } from "./api";
import { hasSavedViews } from "./view-presence";
import type { Preset } from "./types";

jest.mock("./api", () => ({ listPresets: jest.fn() }));
const TEAM = "00000000-0000-4000-8000-000000000001";
const ctx = { workspaceSlug: "acme", session: null };

beforeEach(() => jest.clearAllMocks());

it("follows fresh cursors past mismatched records without counting templates or other teams", async () => {
  jest
    .mocked(listPresets)
    .mockResolvedValueOnce({
      items: [{ kind: "template", teamId: TEAM } as Preset],
      nextCursor: "second",
    })
    .mockResolvedValueOnce({
      items: [{ kind: "view", teamId: "other" } as Preset],
      nextCursor: "third",
    })
    .mockResolvedValueOnce({
      items: [{ kind: "view", teamId: TEAM } as Preset],
      nextCursor: "",
    });
  await expect(
    hasSavedViews(TEAM, ctx, new AbortController().signal),
  ).resolves.toBe(true);
  expect(jest.mocked(listPresets).mock.calls.map((call) => call[2])).toEqual([
    "",
    "second",
    "third",
  ]);
  expect(
    jest.mocked(listPresets).mock.calls.every((call) => call[5] === 1),
  ).toBe(true);
});

it("rejects a repeated cursor instead of looping or claiming a positive result", async () => {
  jest
    .mocked(listPresets)
    .mockResolvedValue({ items: [], nextCursor: "repeat" });
  await expect(
    hasSavedViews(TEAM, ctx, new AbortController().signal),
  ).rejects.toThrow("Views could not be checked");
  expect(listPresets).toHaveBeenCalledTimes(2);
});
