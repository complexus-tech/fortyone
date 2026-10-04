import { act, fireEvent, render, screen } from "@testing-library/react";
import { toast } from "sonner";
import {
  favoritesKey,
  readFavoriteSnapshot,
  updateFavorites,
} from "@/shared/favorites/store";
import {
  getServerFavorites,
  readViewFavorites,
  subscribeViewFavorites,
  updateViewFavorites,
  viewFavoritesKey,
} from "./favorites-store";
import { useFavoriteViews } from "./use-favorite-views";

let mockUserId = "user-one";
let mockWorkspace = "acme";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const VIEW = {
  id: "00000000-0000-4000-8000-000000000001",
  teamId: "00000000-0000-4000-8000-000000000002",
};
const Probe = ({ name }: { name: string }) => {
  const { favorites, toggleFavorite } = useFavoriteViews();
  return (
    <>
      <button
        onClick={() => {
          toggleFavorite(VIEW);
        }}
        type="button"
      >
        {name}
      </button>
      <output aria-label={name}>{favorites.length}</output>
    </>
  );
};

describe("saved-view favorite preferences", () => {
  beforeEach(() => {
    localStorage.clear();
    mockUserId = "user-one";
    mockWorkspace = "acme";
    jest.clearAllMocks();
  });

  it("synchronizes independent consumers and stores only references", () => {
    render(
      <>
        <Probe name="Header" />
        <Probe name="Sidebar" />
      </>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Header" }));
    expect(screen.getByLabelText("Header")).toHaveTextContent("1");
    expect(screen.getByLabelText("Sidebar")).toHaveTextContent("1");
    expect(
      JSON.parse(
        localStorage.getItem(
          favoritesKey({ workspaceSlug: "acme", userId: "user-one" }),
        )!,
      ),
    ).toEqual({ version: 2, favorites: [{ kind: "view", ...VIEW }] });
    fireEvent.click(screen.getByRole("button", { name: "Sidebar" }));
    expect(screen.getByLabelText("Header")).toHaveTextContent("0");
  });

  it("isolates user/workspace scopes immediately and restores each preference", () => {
    const { rerender } = render(<Probe name="Favorites" />);
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    mockUserId = "user-two";
    rerender(<Probe name="Favorites" />);
    expect(screen.getByLabelText("Favorites")).toHaveTextContent("0");
    mockUserId = "user-one";
    mockWorkspace = "other";
    rerender(<Probe name="Favorites" />);
    expect(screen.getByLabelText("Favorites")).toHaveTextContent("0");
    mockWorkspace = "acme";
    rerender(<Probe name="Favorites" />);
    expect(screen.getByLabelText("Favorites")).toHaveTextContent("1");
  });

  it("handles cross-tab storage changes and cleans up subscriptions", () => {
    const key = viewFavoritesKey("acme", "user-one");
    const listener = jest.fn();
    const unsubscribe = subscribeViewFavorites(key, listener);
    render(<Probe name="Favorites" />);
    act(() => {
      localStorage.setItem(key, JSON.stringify([VIEW]));
      window.dispatchEvent(new StorageEvent("storage", { key }));
    });
    expect(screen.getByLabelText("Favorites")).toHaveTextContent("1");
    expect(listener).toHaveBeenCalledTimes(1);
    unsubscribe();
    window.dispatchEvent(new StorageEvent("storage", { key }));
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("rejects malformed references, strips snapshots, deduplicates and keeps snapshots stable", () => {
    const key = viewFavoritesKey("acme", "user-one");
    localStorage.setItem(
      key,
      JSON.stringify([
        { ...VIEW, configuration: { layout: "kanban" }, name: "Old name" },
        VIEW,
      ]),
    );
    const refs = readViewFavorites(key);
    expect(refs).toEqual([VIEW]);
    expect(readViewFavorites(key)).toBe(refs);
    localStorage.setItem(
      key,
      JSON.stringify([{ id: "invalid", teamId: VIEW.teamId }]),
    );
    expect(readViewFavorites(key)).toEqual([]);
    expect(getServerFavorites()).toBe(getServerFavorites());
    expect(readViewFavorites(null)).toEqual([]);
  });

  it("preserves browser state and reports storage failures", () => {
    render(<Probe name="Favorites" />);
    const write = jest
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new Error("Storage unavailable");
      });
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    expect(screen.getByLabelText("Favorites")).toHaveTextContent("0");
    expect(toast.error).toHaveBeenCalledWith(
      "Favorites could not be saved in this browser.",
    );
    write.mockRestore();
  });

  it("reads the latest stored references before applying successive updates", () => {
    const key = viewFavoritesKey("acme", "user-one");
    updateViewFavorites(key, () => [VIEW]);
    updateViewFavorites(key, (current) => [
      ...current,
      { ...VIEW, id: "00000000-0000-4000-8000-000000000003" },
    ]);
    expect(readViewFavorites(key)).toHaveLength(2);
  });

  it("keeps task and team favorites when the legacy view adapter replaces or removes views", () => {
    const scope = { workspaceSlug: "acme", userId: "user-one" };
    const nonViews = [
      { kind: "story" as const, id: VIEW.id },
      { kind: "team" as const, id: VIEW.teamId },
    ];
    updateFavorites(scope, () => nonViews);
    const key = viewFavoritesKey("acme", "user-one");
    updateViewFavorites(key, () => [VIEW]);
    expect(readViewFavorites(key)).toEqual([VIEW]);
    updateViewFavorites(key, () => []);
    expect(readFavoriteSnapshot(scope).favorites).toEqual(nonViews);
  });
});
