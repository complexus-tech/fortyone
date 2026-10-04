import { z } from "zod";

export const favoriteRefSchema = z.discriminatedUnion("kind", [
  z.object({ kind: z.literal("view"), id: z.uuid(), teamId: z.uuid() }),
  z.object({
    kind: z.literal("story"),
    id: z.uuid(),
    teamId: z.uuid().optional(),
  }),
  z.object({ kind: z.literal("team"), id: z.uuid() }),
]);

export type FavoriteRef = z.infer<typeof favoriteRefSchema>;
export type FavoriteScope = { workspaceSlug: string; userId: string };
export const favoriteIdentity = (item: FavoriteRef) =>
  `${item.kind}:${item.id}`;
export const sameFavorite = (left: FavoriteRef, right: FavoriteRef) =>
  favoriteIdentity(left) === favoriteIdentity(right);
