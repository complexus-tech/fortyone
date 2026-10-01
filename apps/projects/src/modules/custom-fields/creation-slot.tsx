"use client";

import { useImperativeHandle } from "react";
import type { CreationPropertiesProps } from "@/shared/story/creation-property-slots";
import { CreateCustomFields, useCreateCustomFields } from "./create-fields";

export const CustomFieldsCreationSlot = ({
  controllerRef,
  teamId,
  disabled,
}: CreationPropertiesProps) => {
  const fields = useCreateCustomFields(teamId);
  useImperativeHandle(controllerRef, () => fields, [fields]);
  return (
    <CreateCustomFields
      disabled={disabled}
      onChange={fields.setValues}
      teamId={teamId}
      values={fields.values}
    />
  );
};
