"use client";

import { FavoriteButton } from "@/shared/favorites";
import type { FavoriteViewRef } from "./favorites-store";

export const ViewFavoriteButton = ({
  view,
  name,
}: {
  view: FavoriteViewRef;
  name: string;
}) => <FavoriteButton item={{ kind: "view", ...view }} name={name} />;
