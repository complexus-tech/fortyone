"use client";

import { useQueries } from "@tanstack/react-query";
import { usePathname, useSearchParams } from "next/navigation";
import { Box, Button, Flex, Text, Tooltip } from "ui";
import { ViewsIcon, CloseIcon } from "icons";
import { cn } from "lib";
import { NavLink } from "@/components/ui/nav-link";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useFavoriteViews } from "./use-favorite-views";
import { presetKey } from "./hooks";
import { resolveSavedViews } from "./resolve-views";
import { savedViewPath } from "./view-link";
import { selectCurrentView } from "./view-selection";

export const ViewFavoritesSidebar = ({
  isCollapsed,
}: {
  isCollapsed: boolean;
}) => {
  const { favorites, removeFavorite } = useFavoriteViews();
  const { data: session } = useSession();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const pathname = usePathname();
  const viewId = useSearchParams().get("view");
  const teamIds = [...new Set(favorites.map((favorite) => favorite.teamId))];
  const queries = useQueries({
    queries: teamIds.map((teamId) => ({
      queryKey: [
        ...presetKey(workspaceSlug, session?.user.id ?? "", teamId, "view"),
        "favorites",
        favorites
          .filter((favorite) => favorite.teamId === teamId)
          .map((favorite) => favorite.id),
      ],
      queryFn: ({ signal }: { signal: AbortSignal }) =>
        resolveSavedViews(
          teamId,
          favorites
            .filter((favorite) => favorite.teamId === teamId)
            .map((favorite) => favorite.id),
          { session, workspaceSlug },
          signal,
        ),
      enabled: Boolean(session),
      staleTime: 60_000,
    })),
  });
  if (!favorites.length) return null;
  return (
    <Box className={cn("mt-4", isCollapsed && "mt-3")}>
      <Text
        as="h2"
        className={cn("mb-2 pl-2.5 font-medium", isCollapsed && "sr-only")}
        color="muted"
      >
        Favorites
      </Text>
      <Flex aria-label="Favorite views" direction="column" gap={1} role="group">
        {favorites.map((favorite) => {
          const query = queries[teamIds.indexOf(favorite.teamId)];
          const view = !query.isError
            ? query.data?.find((item) => item.id === favorite.id)
            : undefined;
          const unavailable = query.isSuccess && !view;
          let name = "View unavailable";
          if (view) name = view.name;
          else if (query.isPending) name = "Loading view...";
          else if (query.isError) name = "View could not be loaded";
          const path = withWorkspace(`/teams/${favorite.teamId}/stories`);
          const active = pathname === path && viewId === favorite.id;
          const content = (
            <>
              <ViewsIcon className={cn("shrink-0", isCollapsed && "h-5.5")} />
              <span
                className={cn(
                  "min-w-0 flex-1 truncate",
                  isCollapsed &&
                    "line-clamp-2 w-full flex-none text-xs leading-3.5 font-semibold whitespace-normal",
                )}
              >
                {name}
              </span>
            </>
          );
          const classes = cn(
            "hover:bg-primary/5 hover:text-primary hover:[&_svg]:text-primary relative pr-8",
            active && "bg-primary/5 text-primary [&_svg]:text-primary",
            isCollapsed &&
              "min-h-14 flex-col justify-center gap-1 px-1 py-2 text-center",
          );
          return (
            <Box
              className="group relative"
              key={`${favorite.teamId}:${favorite.id}`}
            >
              {unavailable ? (
                <Flex
                  aria-label={name}
                  className={cn(classes, "text-text-muted px-2 py-2")}
                  gap={2}
                >
                  {content}
                </Flex>
              ) : (
                <NavLink
                  active={active}
                  aria-current={active ? "page" : undefined}
                  aria-label={name}
                  className={classes}
                  href={withWorkspace(
                    savedViewPath(favorite.teamId, favorite.id),
                  )}
                  onClick={(event) => {
                    if (
                      !active ||
                      !session ||
                      event.metaKey ||
                      event.ctrlKey ||
                      event.shiftKey ||
                      event.altKey
                    )
                      return;
                    event.preventDefault();
                    selectCurrentView({
                      ...favorite,
                      workspaceSlug,
                      userId: session.user.id,
                    });
                  }}
                  title={name}
                >
                  {content}
                </NavLink>
              )}
              <Tooltip title={`Remove ${name} from favorites`}>
                <Button
                  aria-label={`Remove ${name} from favorites`}
                  asIcon
                  className={cn(
                    "absolute top-1 right-0 opacity-100 group-hover:opacity-100 focus-visible:opacity-100 [@media(hover:hover)]:opacity-0",
                    isCollapsed && "top-0 right-0 h-5 w-5",
                  )}
                  color="tertiary"
                  onClick={() => {
                    removeFavorite(favorite);
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
      </Flex>
    </Box>
  );
};
