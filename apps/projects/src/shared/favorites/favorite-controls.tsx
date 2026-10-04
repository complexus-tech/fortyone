"use client";

import { StarIcon } from "icons";
import { Button, ContextMenu, Menu, Tooltip } from "ui";
import { cn } from "lib";
import type { FavoriteRef } from "./types";
import { useFavorites } from "./use-favorites";

type FavoriteProps = { item: FavoriteRef; name: string; disabled?: boolean };
const useFavoriteAction = ({ item, name, disabled = false }: FavoriteProps) => {
  const { enabled, isFavorite, toggleFavorite } = useFavorites();
  const favorite = isFavorite(item);
  const unavailable = disabled || !enabled;
  return {
    favorite,
    disabled: unavailable,
    label: `${favorite ? "Remove" : "Add"} ${name} ${favorite ? "from" : "to"} favorites`,
    toggle: () => {
      if (!unavailable) toggleFavorite(item);
    },
  };
};
const FavoriteStar = ({ favorite }: { favorite: boolean }) => (
  <StarIcon
    aria-hidden
    className={favorite ? "text-primary h-4.5" : "h-4.5 [&_path]:fill-none"}
  />
);
export const FavoriteButton = ({
  className,
  ...props
}: FavoriteProps & { className?: string }) => {
  const action = useFavoriteAction(props);
  return (
    <Tooltip title={action.label}>
      <Button
        aria-label={action.label}
        aria-pressed={action.favorite}
        asIcon
        className={cn("shrink-0", className)}
        color="tertiary"
        disabled={action.disabled}
        onClick={action.toggle}
        size="sm"
        type="button"
        variant="naked"
      >
        <FavoriteStar favorite={action.favorite} />
      </Button>
    </Tooltip>
  );
};
export const FavoriteMenuItem = (props: FavoriteProps) => {
  const action = useFavoriteAction(props);
  return (
    <Menu.Item
      aria-label={action.label}
      disabled={action.disabled}
      onSelect={action.toggle}
    >
      <FavoriteStar favorite={action.favorite} />
      {action.favorite ? "Remove favorite" : "Favorite"}
    </Menu.Item>
  );
};
export const FavoriteContextMenuItem = (props: FavoriteProps) => {
  const action = useFavoriteAction(props);
  return (
    <ContextMenu.Item
      aria-label={action.label}
      disabled={action.disabled}
      onSelect={action.toggle}
    >
      <FavoriteStar favorite={action.favorite} />
      {action.favorite ? "Remove favorite" : "Favorite"}
    </ContextMenu.Item>
  );
};
