import type { useMembers } from "@/lib/hooks/members";

export type KeyResultsMember = NonNullable<
  ReturnType<typeof useMembers>["data"]
>[number];
