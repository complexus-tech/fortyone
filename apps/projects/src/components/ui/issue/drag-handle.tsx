import { GripVertical } from "lucide-react";

export const DragHandle = () => {
  return (
    <GripVertical className="absolute -left-[2.9rem] hidden h-5 w-auto cursor-move text-gray group-hover:inline-block" />
  );
};
