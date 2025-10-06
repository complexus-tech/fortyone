import { useViewOptions as useBaseViewOptions } from "@/hooks/use-view-options";

export const useViewOptions = () => {
  return useBaseViewOptions("my-work:view-options");
};
