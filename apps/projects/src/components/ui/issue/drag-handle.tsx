import { DragIcon } from "icons";
import type { Icon } from "icons/src/types";

export const DragHandle = (props: Icon) => {
  return (
    <DragIcon
      className="absolute -left-[2.9rem] hidden h-5 w-auto cursor-move text-gray group-hover:inline-block"
      {...props}
    />
  );
};
