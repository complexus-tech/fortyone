"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { usePathname, useSearchParams } from "next/navigation";
import { Box, Button, Collapsible, Flex, Text, Tooltip } from "ui";
import { ChevronRightIcon, CloseIcon, StoryIcon } from "icons";
import { cn } from "lib";
import { NavLink } from "@/components/ui/nav-link";
import { TeamColor } from "@/components/ui/team-color";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getStatuses } from "@/lib/queries/states/get-states";
import { statusKeys } from "@/constants/keys";
import { favoriteIdentity, useFavorites } from "@/shared/favorites";
import type { FavoriteRef } from "@/shared/favorites";
import { SavedViewIcon } from "@/shared/views/icons";
import { selectCurrentView } from "@/modules/work-presets/public/views";
import type { ResolvedFavorite } from "./favorites-resolution";
import { favoriteRequests, loadFavoriteSource } from "./favorites-resolution";

const subscribeToHydration = () => () => {};
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;
const readExpansion = (key: string) => {
  try {
    const raw = window.localStorage.getItem(key);
    const stored: unknown = raw === null ? true : JSON.parse(raw);
    return typeof stored === "boolean" ? stored : true;
  } catch {
    return true;
  }
};

/** Match local-storage hydration and scope resets while optional expansion remains usable without storage. */
const useFavoritesExpansion = (workspaceSlug: string, userId: string) => {
  const key = `sidebar:${encodeURIComponent(workspaceSlug)}:${encodeURIComponent(userId)}:favorites-expanded`;
  const hydrated = useSyncExternalStore(
    subscribeToHydration,
    getClientSnapshot,
    getServerSnapshot,
  );
  const [stored, setStored] = useState(() => ({
    key,
    hydrated,
    expanded: hydrated ? readExpansion(key) : true,
  }));
  let current = stored;
  if (stored.key !== key || (!stored.hydrated && hydrated)) {
    current = { key, hydrated, expanded: hydrated ? readExpansion(key) : true };
    setStored(current);
  }
  const setExpanded = (expanded: boolean) => {
    setStored({ key, hydrated: true, expanded });
    try {
      window.localStorage.setItem(key, JSON.stringify(expanded));
    } catch {
      // Keep the section usable in memory when browser persistence is blocked.
    }
  };
  return [current.expanded, setExpanded] as const;
};

export const FavoritesSidebar = ({ isCollapsed }: { isCollapsed: boolean }) => {
  const { favorites, error, removeFavorites } = useFavorites();
  const container = useRef<HTMLDivElement>(null);
  const pendingFocus = useRef<{
    next: HTMLElement | null;
    fallback: HTMLElement | null;
  } | null>(null);
  const { data: session } = useSession();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const pathname = usePathname();
  const viewId = useSearchParams().get("view");
  const userId = session?.user.id ?? "";
  const [expanded, setExpanded] = useFavoritesExpansion(workspaceSlug, userId);
  const scope = `${workspaceSlug}:${userId}`;
  const currentScope = useRef(scope);
  useLayoutEffect(() => {
    currentScope.current = scope;
  }, [scope]);
  const ctx = { session, workspaceSlug };
  const requests = favoriteRequests(favorites, userId, ctx);
  const queries = useQueries({
    queries: requests.map((request) => ({
      queryKey: request.key,
      queryFn: ({ signal }: { signal: AbortSignal }) =>
        loadFavoriteSource(request, signal),
      select: request.project,
      enabled: Boolean(userId && workspaceSlug),
      staleTime: 60_000,
      refetchOnMount: "always" as const,
      retry: 1,
    })),
  });
  const statuses = useQuery({
    queryKey: [...statusKeys.lists(workspaceSlug), "favorite", userId],
    queryFn: () => getStatuses(ctx),
    enabled: Boolean(userId && favorites.some((item) => item.kind === "story")),
    staleTime: 60_000,
    refetchOnMount: "always",
  });
  const resolved = new Map<string, ResolvedFavorite>();
  const unavailable = requests.flatMap((request, index) => {
    const query = queries[index];
    if (query.isError || !query.isFetchedAfterMount) return [];
    const items = query.data ?? [];
    items.forEach((item) => resolved.set(favoriteIdentity(item.ref), item));
    return request.refs.filter(
      (ref) =>
        !items.some(
          (item) => favoriteIdentity(item.ref) === favoriteIdentity(ref),
        ),
    );
  });
  const loading = queries.some(
    (query) => !query.isFetchedAfterMount && !query.isError,
  );
  const failed = queries.filter((query) => query.isError);
  useLayoutEffect(() => {
    const target = pendingFocus.current?.next?.isConnected
      ? pendingFocus.current.next
      : pendingFocus.current?.fallback;
    if (target?.isConnected) target.focus();
    pendingFocus.current = null;
  }, [favorites]);
  const remove = (items: readonly FavoriteRef[]) => {
    const removing = new Set(items.map(favoriteIdentity));
    const links = Array.from(
      container.current?.querySelectorAll<HTMLAnchorElement>(
        "a[data-favorite-key]",
      ) ?? [],
    );
    const index = links.findIndex((link) =>
      removing.has(link.dataset.favoriteKey ?? ""),
    );
    const remaining = links.filter(
      (link) => !removing.has(link.dataset.favoriteKey ?? ""),
    );
    const next =
      remaining.find((link) => links.indexOf(link) > index) ?? remaining.at(-1);
    const fallback = container.current
      ?.closest("[data-sidebar-content]")
      ?.querySelector<HTMLElement>("[data-workspace-switcher], a[href]");
    if (removeFavorites(items))
      pendingFocus.current = { next: next ?? null, fallback: fallback ?? null };
  };
  const retry = async (button: HTMLButtonElement) => {
    await Promise.all(failed.map((query) => query.refetch()));
    requestAnimationFrame(() => {
      // Recovery can remove the retry control. Keep focus local, and never reclaim it after the user moves on.
      if (
        currentScope.current !== scope ||
        button.isConnected ||
        document.activeElement !== document.body
      )
        return;
      const target =
        container.current?.querySelector<HTMLElement>("a[data-favorite-key]") ??
        container.current
          ?.closest("[data-sidebar-content]")
          ?.querySelector<HTMLElement>("[data-workspace-switcher], a[href]");
      target?.focus();
    });
  };
  if (!favorites.length && !error) return null;
  const icon = (item: ResolvedFavorite) => {
    const className = cn("size-4 shrink-0", isCollapsed && "size-5");
    if (item.icon.kind === "team")
      return <TeamColor className={className} color={item.icon.color} />;
    if (item.icon.kind === "view")
      return <SavedViewIcon className={className} configuration={item.icon} />;
    const statusId = item.icon.statusId;
    const status =
      !statuses.isError && statuses.isFetchedAfterMount
        ? statuses.data?.find((state) => state.id === statusId)
        : undefined;
    return (
      <StoryIcon
        aria-hidden
        className={className}
        style={status ? { color: status.color } : undefined}
      />
    );
  };
  return (
    <div className={cn("mt-4", isCollapsed && "mt-3")} ref={container}>
      <Collapsible onOpenChange={setExpanded} open={isCollapsed || expanded}>
        <Text as="h2" className={isCollapsed ? "sr-only" : undefined}>
          {isCollapsed ? (
            "Favorites"
          ) : (
            <Collapsible.Trigger asChild>
              <button
                aria-label="Favorites"
                className="text-text-muted focus-visible:ring-primary/40 hover:text-foreground flex h-8 items-center gap-1 rounded-lg px-2.5 text-left font-medium transition-colors outline-none focus-visible:ring-2"
                type="button"
              >
                <span>Favorites</span>
                <ChevronRightIcon
                  aria-hidden
                  className={cn(
                    "h-3.5 w-auto transition-transform duration-200",
                    expanded && "rotate-90",
                  )}
                  strokeWidth={3.5}
                />
              </button>
            </Collapsible.Trigger>
          )}
        </Text>
        <Collapsible.Content>
          <Flex
            aria-label="Favorites"
            className={cn("mt-1", isCollapsed && "mt-0")}
            direction="column"
            gap={1}
            role="group"
          >
            {favorites.map((ref) => {
              const item = resolved.get(favoriteIdentity(ref));
              if (!item) return null;
              const href = withWorkspace(item.path);
              const active =
                ref.kind === "view"
                  ? pathname ===
                      withWorkspace(`/teams/${ref.teamId}/stories`) &&
                    viewId === ref.id
                  : (pathname === href ||
                      (ref.kind === "story" &&
                        pathname === withWorkspace(`/work/${ref.id}`))) &&
                    !viewId;
              return (
                <Box className="group relative" key={favoriteIdentity(ref)}>
                  <NavLink
                    active={active}
                    aria-current={active ? "page" : undefined}
                    aria-label={item.name}
                    className={cn(
                      "hover:bg-primary/5 hover:text-primary focus-visible:ring-ring hover:[&_svg]:text-primary relative pr-8 focus-visible:ring-2",
                      active &&
                        "bg-primary/5 text-primary [&_svg]:text-primary",
                      isCollapsed &&
                        "min-h-14 flex-col justify-center gap-1 px-1 py-2 text-center",
                    )}
                    data-favorite-key={favoriteIdentity(ref)}
                    href={href}
                    onClick={(event) => {
                      if (
                        !active ||
                        ref.kind !== "view" ||
                        !userId ||
                        event.metaKey ||
                        event.ctrlKey ||
                        event.shiftKey ||
                        event.altKey
                      )
                        return;
                      event.preventDefault();
                      selectCurrentView({
                        id: ref.id,
                        teamId: ref.teamId,
                        workspaceSlug,
                        userId,
                      });
                    }}
                    title={item.name}
                  >
                    {icon(item)}
                    <span
                      className={cn(
                        "min-w-0 flex-1 truncate",
                        isCollapsed &&
                          "line-clamp-2 w-full flex-none text-xs leading-3.5 font-semibold whitespace-normal",
                      )}
                    >
                      {item.name}
                    </span>
                  </NavLink>
                  <Tooltip title={`Remove ${item.name} from favorites`}>
                    <Button
                      aria-label={`Remove ${item.name} from favorites`}
                      asIcon
                      className={cn(
                        "absolute top-1 right-0 opacity-100 group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:opacity-100 [@media(hover:hover)]:opacity-0",
                        isCollapsed && "top-0 right-0 h-5 w-5",
                      )}
                      color="tertiary"
                      onClick={() => {
                        remove([ref]);
                      }}
                      size="sm"
                      type="button"
                      variant="naked"
                    >
                      <CloseIcon className="h-3.5 w-auto" />
                    </Button>
                  </Tooltip>
                </Box>
              );
            })}
            {loading ? (
              <Text aria-live="polite" className="px-2.5 py-1" color="muted">
                Loading favorites...
              </Text>
            ) : null}
          </Flex>
        </Collapsible.Content>
      </Collapsible>
      {error ? (
        <Text className="px-2.5 py-1" color="danger" role="alert">
          {error}
        </Text>
      ) : null}
      {failed.length ? (
        <Box className="px-2.5 py-1">
          <Text color="danger" role="alert">
            Some favorites could not be loaded.
          </Text>
          <Button
            color="tertiary"
            disabled={failed.some((query) => query.isFetching)}
            onClick={(event) => {
              void retry(event.currentTarget);
            }}
            size="sm"
            variant="naked"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      {unavailable.length ? (
        <Box className="px-2.5 py-1">
          <Text color="muted" role="status">
            Some favorites are no longer available.
          </Text>
          <Button
            color="tertiary"
            onClick={() => {
              remove(unavailable);
            }}
            size="sm"
            variant="naked"
          >
            Remove unavailable favorites
          </Button>
        </Box>
      ) : null}
    </div>
  );
};
