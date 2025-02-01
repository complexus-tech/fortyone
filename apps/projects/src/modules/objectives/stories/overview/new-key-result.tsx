import { useParams } from "next/navigation";
import React from "react";

export const NewKeyResult = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  return <div />;
};
