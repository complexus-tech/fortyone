import { Dot } from "@/components/ui/dot";
import { hexToRgba } from "@/utils";
import type { PublicRequestBoard } from "./types";

const getBoardPillStyle = (color?: string) => {
  if (!color || !/^#?[0-9a-f]{6}$/i.test(color)) return undefined;

  return {
    backgroundColor: hexToRgba(color, 0.1),
    borderColor: hexToRgba(color, 0.2),
  };
};

export const PublicBoardPill = ({ board }: { board: PublicRequestBoard }) => (
  <span
    className="border-border bg-surface-muted/40 inline-flex h-7 min-w-0 items-center gap-1.5 rounded-md border px-2 text-[0.9rem] font-medium"
    style={getBoardPillStyle(board.color)}
  >
    <Dot color={board.color ?? "var(--color-text-muted)"} />
    <span className="truncate">{board.name}</span>
  </span>
);
