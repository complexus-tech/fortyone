"use client";

import type { ComponentType, ReactNode } from "react";
import { createContext, createElement, useContext } from "react";

type FavoritesProps = { isCollapsed: boolean };
const FavoritesContext = createContext<ComponentType<FavoritesProps> | null>(
  null,
);

export const ViewFavoritesProvider = ({
  children,
  Favorites,
}: {
  children: ReactNode;
  Favorites: ComponentType<FavoritesProps>;
}) => (
  <FavoritesContext.Provider value={Favorites}>
    {children}
  </FavoritesContext.Provider>
);

export const ViewFavoritesSlot = (props: FavoritesProps) => {
  const Favorites = useContext(FavoritesContext);
  return Favorites ? createElement(Favorites, props) : null;
};
