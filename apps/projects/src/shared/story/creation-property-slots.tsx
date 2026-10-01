"use client";

import { createContext, useContext, useImperativeHandle } from "react";
import type { ComponentType, ReactNode, Ref } from "react";
import type { NewStory } from "./types";
import type { TaskTemplateConfiguration } from "./task-template";

export type CreationPropertiesController = {
  isPending: boolean;
  isError: boolean;
  prepareValues: () => NonNullable<NewStory["customFieldValues"]>;
  reset: () => void;
  setValues: (values: Record<string, string | null>) => void;
};

export type CreationPropertiesProps = {
  controllerRef: Ref<CreationPropertiesController>;
  teamId?: string;
  disabled?: boolean;
};

export type CreationTemplatePickerProps = {
  teamId: string;
  disabled?: boolean;
  onSelect: (template: TaskTemplateConfiguration) => void;
};

export type CreationPropertySlots = {
  Properties: ComponentType<CreationPropertiesProps>;
  TemplatePicker: ComponentType<CreationTemplatePickerProps>;
};

export const EMPTY_CREATION_PROPERTIES: CreationPropertiesController = {
  isPending: false,
  isError: false,
  prepareValues: () => [],
  reset: () => undefined,
  setValues: () => undefined,
};

const EmptyProperties = ({ controllerRef }: CreationPropertiesProps) => {
  useImperativeHandle(controllerRef, () => EMPTY_CREATION_PROPERTIES, []);
  return null;
};

const EmptyTemplatePicker = () => null;
const DEFAULT_SLOTS: CreationPropertySlots = {
  Properties: EmptyProperties,
  TemplatePicker: EmptyTemplatePicker,
};

const CreationPropertyContext = createContext(DEFAULT_SLOTS);

/** Shell supplies feature implementations; shared controls own no feature queries. */
export const CreationPropertySlotsProvider = ({
  children,
  slots,
}: {
  children: ReactNode;
  slots: CreationPropertySlots;
}) => (
  <CreationPropertyContext.Provider value={slots}>
    {children}
  </CreationPropertyContext.Provider>
);

export const useCreationPropertySlots = () =>
  useContext(CreationPropertyContext);
