"use client";

import { useEffect, useRef, useState } from "react";
import { parseAsString, useQueryState } from "nuqs";
import { useRouter } from "next/navigation";
import { Button, Flex, Skeleton } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { SavedViewLoadState } from "@/shared/story/view-configuration";
import type { SavedViewConfiguration } from "./types";
import { useSelectedView } from "./use-selected-view";
import { SavedViewIdentity } from "./saved-view-identity";
import { ViewsSwitcher } from "./views-switcher";
import { savedViewPath } from "./view-link";
import {
  consumeCreatedViewFocus,
  subscribeViewSelection,
} from "./view-selection";

const selectedViewStatus = (
  viewId: string | null,
  query: ReturnType<typeof useSelectedView>,
): SavedViewLoadState["status"] => {
  if (!viewId) return "idle";
  if (query.isError) return "error";
  if (query.isFetchedAfterMount && query.data) return "ready";
  if (!query.isFetchedAfterMount || query.isFetching) return "loading";
  return "unavailable";
};

export const SavedViews = ({
  teamId,
  onApply,
  onLoadStateChange,
}: {
  teamId: string;
  configuration: SavedViewConfiguration;
  onApply: (configuration: SavedViewConfiguration) => void;
  onLoadStateChange?: (state: SavedViewLoadState) => void;
}) => {
  const router = useRouter();
  const [selection, setSelection] = useState(0);
  const trigger = useRef<HTMLButtonElement>(null);
  const { data: session } = useSession();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const [viewId, setViewId] = useQueryState("view", parseAsString);
  const query = useSelectedView(teamId, viewId, selection);
  const applied = useRef<string | null>(null);
  const published = useRef<string | null>(null);
  const scope = `${workspaceSlug}:${session?.user.id ?? ""}:${teamId}:${viewId ?? ""}:${selection}`;
  const status = selectedViewStatus(viewId, query);
  const selected = status === "ready" ? query.data ?? null : null;
  const reselect = () => {
    onLoadStateChange?.({
      workspaceSlug,
      userId: session?.user.id ?? "",
      teamId,
      viewId,
      status: "loading",
    });
    setSelection((current) => current + 1);
  };
  useEffect(() => {
    if (!viewId || !session) return;
    return subscribeViewSelection(
      { workspaceSlug, userId: session.user.id, teamId, id: viewId },
      () => {
        onLoadStateChange?.({
          workspaceSlug,
          userId: session.user.id,
          teamId,
          viewId,
          status: "loading",
        });
        setSelection((current) => current + 1);
      },
    );
  }, [workspaceSlug, session, teamId, viewId, onLoadStateChange]);
  useEffect(() => {
    const application = scope;
    if (status === "ready" && query.data && applied.current !== application) {
      applied.current = application;
      onApply(query.data.configuration);
    }
    const publication = `${scope}:${status}`;
    if (
      viewId &&
      status !== "loading" &&
      status !== "idle" &&
      consumeCreatedViewFocus({
        workspaceSlug,
        userId: session?.user.id ?? "",
        teamId,
        id: viewId,
      })
    )
      requestAnimationFrame(() => {
        const expectedScope = `${workspaceSlug}:${session?.user.id ?? ""}:${teamId}`;
        const header = trigger.current?.closest<HTMLElement>(
          "[data-view-focus-scope]",
        );
        if (
          header?.dataset.viewFocusScope !== expectedScope ||
          header.dataset.viewId !== viewId
        )
          return;
        const activeScope = document.activeElement?.closest<HTMLElement>(
          "[data-view-focus-scope]",
        )?.dataset.viewFocusScope;
        if (
          document.activeElement === document.body ||
          activeScope === expectedScope
        )
          trigger.current?.focus();
      });
    if (published.current !== publication) {
      published.current = publication;
      onLoadStateChange?.({
        workspaceSlug,
        userId: session?.user.id ?? "",
        teamId,
        viewId,
        status,
      });
    }
  }, [
    scope,
    status,
    query.data,
    onApply,
    onLoadStateChange,
    teamId,
    viewId,
    workspaceSlug,
    session?.user.id,
  ]);
  const selectView = (preset: { id: string; teamId?: string }) => {
    if (preset.teamId && preset.teamId !== teamId)
      router.push(withWorkspace(savedViewPath(preset.teamId, preset.id)));
    else if (preset.id === viewId) reselect();
    else void setViewId(preset.id);
  };
  if (!viewId) return null;
  return (
    <Flex
      align="center"
      className="min-w-0"
      data-view-focus-scope={`${workspaceSlug}:${session?.user.id ?? ""}:${teamId}`}
      data-view-id={viewId}
      gap={1}
    >
      <ViewsSwitcher
        loading={status === "loading"}
        onSelect={selectView}
        selected={selected}
        triggerRef={selected ? undefined : trigger}
      />
      {selected ? (
        <SavedViewIdentity
          onArchived={() =>
            requestAnimationFrame(() => {
              const header = document.querySelector<HTMLElement>(
                "[data-view-focus-scope]",
              );
              if (
                header?.dataset.viewFocusScope ===
                  `${workspaceSlug}:${session?.user.id ?? ""}:${teamId}` &&
                header.dataset.viewId === viewId &&
                document.activeElement === document.body
              )
                trigger.current?.focus();
            })
          }
          onSelect={selectView}
          selectionControl={false}
          triggerRef={trigger}
          view={selected}
        />
      ) : null}
      {status === "loading" ? (
        <Flex align="center" aria-hidden gap={2}>
          <Skeleton className="size-8 rounded-full" />
          <Skeleton className="size-8 rounded-full" />
        </Flex>
      ) : null}
      {status === "error" || status === "unavailable" ? (
        <Button
          color="tertiary"
          onClick={() => void query.refetch()}
          size="sm"
          variant="outline"
        >
          Try again
        </Button>
      ) : null}
    </Flex>
  );
};
