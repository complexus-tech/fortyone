"use client";

import { useState } from "react";
import { Avatar, Box, Button, Flex, Input, Text } from "ui";
import { SearchIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ViewCatalogTeam } from "./use-view-catalog";
import { useViewCatalog } from "./use-view-catalog";
import { ViewRow } from "./view-row";
import type { SavedView } from "./resolve-views";

type CatalogProps = {
  teams: ViewCatalogTeam[];
  owners?: Record<string, string>;
  teamsPending: boolean;
  teamsError: boolean;
  retryTeams: () => void;
  hasMoreTeams: boolean;
  loadingMoreTeams: boolean;
  loadMoreTeams: () => void;
};

const ViewGroup = ({
  title,
  description,
  avatar,
  rows,
  emptyText,
  onArchived,
}: {
  title: string;
  description?: string;
  avatar?: { name?: string; image?: string | null };
  rows: { view: SavedView; detail: string }[];
  emptyText?: string;
  onArchived: () => void;
}) => (
  <Box>
    <Flex align="center" className="mb-3 min-w-0 gap-2">
      {avatar ? (
        <Avatar name={avatar.name} size="xs" src={avatar.image} />
      ) : null}
      <Text as="h2" fontWeight="medium">
        {title}
      </Text>
      {description ? (
        <Text className="hidden truncate sm:block" color="muted">
          {description}
        </Text>
      ) : null}
    </Flex>
    {rows.map(({ view, detail }) => (
      <ViewRow
        detail={detail}
        key={view.id}
        onArchived={onArchived}
        view={view}
      />
    ))}
    {emptyText && !rows.length ? (
      <Text className="px-3 py-2" color="muted">
        {emptyText}
      </Text>
    ) : null}
  </Box>
);

const WorkspaceViewsCatalog = ({
  teams,
  owners = {},
  teamsPending,
  teamsError,
  retryTeams,
  hasMoreTeams,
  loadingMoreTeams,
  loadMoreTeams,
}: CatalogProps) => {
  const [search, setSearch] = useState("");
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const scope = `${workspaceSlug}:${session?.user.id ?? ""}`;
  const { feeds, userId } = useViewCatalog(teams);
  const matches = (name: string) =>
    name.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase());
  const personal = feeds.flatMap(({ views, team }) =>
    views
      .filter(
        (view) => view.visibility === "personal" && view.ownerId === userId,
      )
      .map((view) => ({ view, team })),
  );
  const visiblePersonal = personal.filter(({ view }) => matches(view.name));
  const shared = feeds.flatMap(({ views, team }) =>
    views
      .filter((view) => view.visibility === "team")
      .map((view) => ({ view, team })),
  );
  const visibleShared = shared.filter(
    ({ view, team }) => matches(view.name) || matches(team.name),
  );
  const loading = teamsPending || feeds.some((feed) => feed.isPending);
  const complete =
    !loading &&
    !teamsError &&
    !hasMoreTeams &&
    !feeds.some((feed) => feed.isError || feed.hasMore || feed.isFetching);
  const emptyPersonal = search
    ? "No matching personal views."
    : "No personal views yet.";
  const emptyShared = search
    ? "No matching team views."
    : "No shared views yet.";
  const returnArchiveFocus = () =>
    requestAnimationFrame(() => {
      const searchInput = document.querySelector<HTMLInputElement>(
        "[data-view-catalog-scope]",
      );
      if (
        searchInput?.dataset.viewCatalogScope === scope &&
        document.activeElement === document.body
      )
        searchInput.focus();
    });
  return (
    <Box className="mx-auto w-full max-w-5xl space-y-6 p-6">
      <Input
        aria-label="Search views"
        className="h-[2.8rem]"
        data-view-catalog-scope={scope}
        leftIcon={<SearchIcon />}
        onChange={(event) => {
          setSearch(event.target.value);
        }}
        placeholder="Search views..."
        value={search}
      />
      {teamsError ? (
        <Box className="space-y-2">
          <Text color="danger" role="alert">
            Teams could not be loaded.
          </Text>
          <Button color="tertiary" onClick={retryTeams} variant="outline">
            Try again
          </Button>
        </Box>
      ) : null}
      {loading ? (
        <Text aria-live="polite" color="muted">
          Loading views...
        </Text>
      ) : null}
      <ViewGroup
        avatar={{ name: session?.user.name, image: session?.user.image }}
        description="Only visible to you"
        emptyText={complete ? emptyPersonal : undefined}
        onArchived={returnArchiveFocus}
        rows={visiblePersonal.map(({ view, team }) => ({
          view,
          detail: team.name,
        }))}
        title="Personal views"
      />
      <ViewGroup
        emptyText={complete ? emptyShared : undefined}
        onArchived={returnArchiveFocus}
        rows={visibleShared.map(({ view, team }) => ({
          view,
          detail: `${team.name} · ${view.ownerId === userId ? "You" : owners[view.ownerId] ?? "Team member"}`,
        }))}
        title="Team views"
      />
      {feeds
        .filter(
          (feed) =>
            feed.isError ||
            feed.hasMore ||
            (feed.isFetching && !feed.isPending),
        )
        .map((feed) => (
          <Box className="space-y-2" key={feed.team.id}>
            {feed.isError ? (
              <>
                <Text color="danger" role="alert">
                  Views from {feed.team.name} could not be loaded.
                </Text>
                <Button
                  color="tertiary"
                  disabled={feed.isFetching}
                  onClick={() => void feed.retry()}
                  variant="outline"
                >
                  Try {feed.team.name} again
                </Button>
              </>
            ) : (
              <Button
                color="tertiary"
                loading={feed.isFetching}
                onClick={feed.loadMore}
                variant="outline"
              >
                Load more from {feed.team.name}
              </Button>
            )}
          </Box>
        ))}
      {hasMoreTeams ? (
        <Button
          color="tertiary"
          loading={loadingMoreTeams}
          onClick={loadMoreTeams}
          variant="outline"
        >
          Load more teams
        </Button>
      ) : null}
      {complete && !personal.length && !shared.length ? (
        <Text color="muted">
          Save filters and display options from a team’s work view to add a view
          here.
        </Text>
      ) : null}
    </Box>
  );
};

export const ViewsCatalog = (props: CatalogProps) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return (
    <WorkspaceViewsCatalog
      key={`${workspaceSlug}:${session?.user.id ?? ""}`}
      {...props}
    />
  );
};
