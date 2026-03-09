import { DragIcon } from "icons";
import type { Icon } from "icons/src/types";

export const DragHandle = (props: Icon) => {
  return (
    <DragIcon
      className="text-text-muted absolute -left-[2.9rem] hidden cursor-move group-hover:md:inline-block"
      strokeWidth={3.5}
      {...props}
    />
  );
};
