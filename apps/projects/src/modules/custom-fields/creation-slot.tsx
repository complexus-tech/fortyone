"use client";

import { useImperativeHandle } from "react";
import type { CreationPropertiesProps } from "@/shared/story/creation-property-slots";
import { CreateCustomFields, useCreateCustomFields } from "./create-fields";

export const CustomFieldsCreationSlot = ({
  controllerRef,
  teamId,
  disabled,
  onValuesChange,
}: CreationPropertiesProps) => {
  const fields = useCreateCustomFields(teamId);
  const setValues = (values: Record<string, string | null>) => {
    fields.setValues(values);
    onValuesChange?.(values);
  };
  useImperativeHandle(controllerRef, () => ({
    ...fields,
    setValues,
    reset: () => {
      fields.reset();
      onValuesChange?.({});
    },
  }));
  return (
    <CreateCustomFields
      disabled={disabled}
      onChange={setValues}
      teamId={teamId}
      values={fields.values}
    />
  );
};
