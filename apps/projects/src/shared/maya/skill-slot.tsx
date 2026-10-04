"use client";

import type { ComponentType, ReactNode } from "react";
import { createContext, createElement, useContext } from "react";

export type MayaSkillSlotProps = {
  value: string;
  onValueChange: (value: string) => void;
  disabled: boolean;
  onReturnFocus: () => void;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

const MayaSkillSlotContext =
  createContext<ComponentType<MayaSkillSlotProps> | null>(null);

/** The shell supplies skill behavior; the shared composer knows only this contract. */
export const MayaSkillSlotProvider = ({
  children,
  Picker,
}: {
  children: ReactNode;
  Picker: ComponentType<MayaSkillSlotProps>;
}) => (
  <MayaSkillSlotContext.Provider value={Picker}>
    {children}
  </MayaSkillSlotContext.Provider>
);

export const useHasMayaSkillSlot = () =>
  Boolean(useContext(MayaSkillSlotContext));

export const MayaSkillSlot = (props: MayaSkillSlotProps) => {
  const Picker = useContext(MayaSkillSlotContext);
  return Picker ? createElement(Picker, props) : null;
};
