import { ApiError } from "api-client";
import type { ViewIconKey } from "@/shared/views/metadata";
import type { WorkspaceCtx } from "@/lib/http";
import type { FavoriteRef } from "@/shared/favorites";
import type { DetailedStory } from "@/shared/story/types";
import type { Team } from "@/modules/teams/public/types";
import type { SavedView } from "@/modules/work-presets/public/views";
import { getStory } from "@/modules/story/public/queries";
import { getTeamDetails } from "@/modules/teams/public/queries";
import {
  presetKey,
  resolveSavedViews,
  savedViewPath,
} from "@/modules/work-presets/public/views";
import { teamKeys } from "@/constants/keys";
import { storyKeys } from "@/shared/story/cache-keys";
import { getStoryPath } from "@/shared/routing/story";

export type ResolvedFavorite = {
  ref: FavoriteRef;
  name: string;
  path: string;
  icon:
    | { kind: "view"; layout: "list" | "kanban"; icon?: ViewIconKey | null }
    | { kind: "story"; statusId: string }
    | { kind: "team"; color: string };
};
type FavoriteSource = DetailedStory | Team[] | SavedView[] | null | undefined;
export type FavoriteRequest = {
  key: readonly unknown[];
  refs: FavoriteRef[];
  load: (signal: AbortSignal) => Promise<FavoriteSource>;
  project: (source: FavoriteSource) => ResolvedFavorite[];
};

export const favoriteIsUnavailable = (error: unknown) =>
  error instanceof ApiError && [403, 404, 410].includes(error.status);

/** Only confirmed missing/denied resources become unavailable; transient failures stay errors. */
export const loadFavoriteSource = async (
  request: FavoriteRequest,
  signal: AbortSignal,
) => {
  try {
    return await request.load(signal);
  } catch (error) {
    if (favoriteIsUnavailable(error)) return null;
    throw error;
  }
};

export const favoriteRequests = (
  favorites: readonly FavoriteRef[],
  userId: string,
  ctx: WorkspaceCtx,
): FavoriteRequest[] => {
  const requests: FavoriteRequest[] = [];
  const views = new Map<string, Extract<FavoriteRef, { kind: "view" }>[]>();
  for (const ref of favorites) {
    if (ref.kind === "view") {
      const teamViews = views.get(ref.teamId) ?? [];
      teamViews.push(ref);
      views.set(ref.teamId, teamViews);
    } else if (ref.kind === "story") {
      requests.push({
        key: [
          ...storyKeys.detail(ctx.workspaceSlug, ref.id),
          "favorite",
          userId,
        ],
        refs: [ref],
        load: () => getStory(ref.id, ctx),
        project: (source) => {
          if (
            !source ||
            Array.isArray(source) ||
            !("title" in source) ||
            source.id !== ref.id ||
            source.deletedAt
          )
            return [];
          return [
            {
              ref,
              name: source.title,
              path: getStoryPath(source),
              icon: { kind: "story", statusId: source.statusId },
            },
          ];
        },
      });
    } else {
      requests.push({
        key: [
          ...teamKeys.lists(ctx.workspaceSlug),
          "favorites",
          userId,
          ref.id,
        ],
        refs: [ref],
        // Membership, deletion and both rename flows invalidate the list prefix.
        load: async (signal) => {
          const team = await getTeamDetails(ref.id, ctx, signal);
          return team ? [team] : [];
        },
        project: (source) => {
          if (!Array.isArray(source)) return [];
          const team = source.find(
            (item): item is Team => "color" in item && item.id === ref.id,
          );
          if (!team) return [];
          return [
            {
              ref,
              name: team.name,
              path: `/teams/${team.id}/stories`,
              icon: { kind: "team", color: team.color },
            },
          ];
        },
      });
    }
  }
  for (const [teamId, refs] of views) {
    requests.push({
      key: [
        ...presetKey(ctx.workspaceSlug, userId, teamId, "view"),
        "favorites",
        refs.map((ref) => ref.id),
      ],
      refs,
      load: (signal) =>
        resolveSavedViews(
          teamId,
          refs.map((ref) => ref.id),
          ctx,
          signal,
        ),
      project: (source) => {
        if (!Array.isArray(source)) return [];
        return refs.flatMap((ref) => {
          const view = source.find(
            (item): item is SavedView =>
              "kind" in item &&
              item.id === ref.id &&
              item.teamId === ref.teamId,
          );
          if (!view) return [];
          return [
            {
              ref,
              name: view.name,
              path: savedViewPath(view.teamId, view.id),
              icon: {
                kind: "view" as const,
                ...(view.configuration.icon
                  ? { icon: view.configuration.icon }
                  : {}),
                layout:
                  view.configuration.layout === "kanban" ? "kanban" : "list",
              },
            },
          ];
        });
      },
    });
  }
  return requests;
};
