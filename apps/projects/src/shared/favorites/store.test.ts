import {
  favoritesKey,
  legacyViewFavoritesKey,
  readFavoriteSnapshot,
  subscribeFavorites,
  updateFavorites,
} from "./store";
import type { FavoriteRef } from "./types";

const scope = { workspaceSlug: "acme", userId: "one" };
const view: FavoriteRef = {
  kind: "view",
  id: "00000000-0000-4000-8000-000000000001",
  teamId: "00000000-0000-4000-8000-000000000002",
};
const story: FavoriteRef = { kind: "story", id: view.id };
const team: FavoriteRef = { kind: "team", id: view.teamId };

describe("generalized favorite references", () => {
  beforeEach(() => {
    localStorage.clear();
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("migrates legacy views on the next write and retains recovery data", () => {
    const legacy = [
      {
        id: view.id,
        teamId: view.teamId,
        name: "Stale name",
        configuration: { layout: "kanban" },
      },
    ];
    localStorage.setItem(legacyViewFavoritesKey(scope), JSON.stringify(legacy));
    expect(readFavoriteSnapshot(scope).favorites).toEqual([view]);
    updateFavorites(scope, (current) => [...current, story, team]);
    expect(JSON.parse(localStorage.getItem(favoritesKey(scope))!)).toEqual({
      version: 2,
      favorites: [view, story, team],
    });
    expect(
      JSON.parse(localStorage.getItem(legacyViewFavoritesKey(scope))!),
    ).toEqual(legacy);
    updateFavorites(scope, () => []);
    expect(readFavoriteSnapshot(scope).favorites).toEqual([]);
  });

  it("isolates both account and workspace and deduplicates by kind and ID", () => {
    updateFavorites(scope, () => [view, story, story, team]);
    expect(readFavoriteSnapshot(scope).favorites).toEqual([view, story, team]);
    expect(readFavoriteSnapshot({ ...scope, userId: "two" }).favorites).toEqual(
      [],
    );
    expect(
      readFavoriteSnapshot({ ...scope, workspaceSlug: "other" }).favorites,
    ).toEqual([]);
  });

  it("never persists names or snapshots and returns stable snapshots until storage changes", () => {
    updateFavorites(scope, () => [
      { ...team, name: "Private title", color: "red" } as FavoriteRef,
    ]);
    const snapshot = readFavoriteSnapshot(scope);
    expect(snapshot.favorites).toEqual([team]);
    expect(readFavoriteSnapshot(scope)).toBe(snapshot);
    expect(localStorage.getItem(favoritesKey(scope))).not.toContain(
      "Private title",
    );
  });

  it("reports malformed preferences and allows an explicit new favorite to recover", () => {
    localStorage.setItem(favoritesKey(scope), "{invalid");
    expect(readFavoriteSnapshot(scope)).toEqual({
      favorites: [],
      error: expect.stringContaining("could not be read"),
    });
    updateFavorites(scope, () => [team]);
    expect(readFavoriteSnapshot(scope)).toEqual({
      favorites: [team],
      error: null,
    });
  });

  it("rejects invalid entity references without replacing existing preferences", () => {
    updateFavorites(scope, () => [team]);
    expect(() => {
      updateFavorites(scope, () => [{ kind: "team", id: "invalid" }]);
    }).toThrow();
    expect(readFavoriteSnapshot(scope).favorites).toEqual([team]);
  });

  it("retains saved state when storage is unreadable or a write fails", () => {
    updateFavorites(scope, () => [team]);
    const write = jest
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new Error("Quota exceeded");
      });
    expect(() => {
      updateFavorites(scope, () => [story]);
    }).toThrow("Quota exceeded");
    expect(readFavoriteSnapshot(scope).favorites).toEqual([team]);
    write.mockRestore();
    jest.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("Storage blocked");
    });
    expect(readFavoriteSnapshot(scope).error).toContain("could not be loaded");
    expect(() => {
      updateFavorites(scope, () => [story]);
    }).toThrow("could not be loaded");
  });

  it("shares listeners, handles cross-tab changes/clear, and stops notifying unmounted consumers", () => {
    const first = jest.fn();
    const second = jest.fn();
    const unsubscribeFirst = subscribeFavorites(scope, first);
    const unsubscribeSecond = subscribeFavorites(scope, second);
    updateFavorites(scope, () => [team]);
    expect(first).toHaveBeenCalledTimes(1);
    expect(second).toHaveBeenCalledTimes(1);
    window.dispatchEvent(
      new StorageEvent("storage", {
        key: favoritesKey({ ...scope, userId: "other" }),
      }),
    );
    expect(first).toHaveBeenCalledTimes(1);
    window.dispatchEvent(new StorageEvent("storage", { key: null }));
    expect(first).toHaveBeenCalledTimes(2);
    unsubscribeFirst();
    window.dispatchEvent(
      new StorageEvent("storage", { key: favoritesKey(scope) }),
    );
    expect(first).toHaveBeenCalledTimes(2);
    expect(second).toHaveBeenCalledTimes(3);
    unsubscribeSecond();
    window.dispatchEvent(new StorageEvent("storage", { key: null }));
    expect(second).toHaveBeenCalledTimes(3);
  });
});
