import type { StoryPriority } from "../stories/types";
import {
  EMPTY_RICH_TEXT,
  isRichTextValue,
  type RichTextValue,
} from "../../components/rich-text/content";

export type SheetName = "team" | "status" | "priority" | "assignee" | "labels";

export const PRIORITIES: StoryPriority[] = [
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
];

export type FormState = {
  title: string;
  description: RichTextValue;
  idempotencyKey: string;
  teamId: string;
  statusId: string | null;
  assigneeId: string | null;
  priority: StoryPriority;
  labelIds: string[];
  activeSheet: SheetName | null;
};

export type FormAction =
  | { type: "setTitle"; title: string }
  | { type: "setDescription"; description: RichTextValue }
  | { type: "setTeam"; teamId: string }
  | { type: "setStatus"; statusId: string }
  | { type: "setAssignee"; assigneeId: string }
  | { type: "setPriority"; priority: StoryPriority }
  | { type: "toggleLabel"; labelId: string }
  | { type: "setSheet"; sheet: SheetName | null };

export const initialState: FormState = {
  title: "",
  description: EMPTY_RICH_TEXT,
  idempotencyKey: "",
  teamId: "",
  statusId: null,
  assigneeId: null,
  priority: "No Priority",
  labelIds: [],
  activeSheet: null,
};

export const isFormState = (value: unknown): value is FormState => {
  if (!value || typeof value !== "object") return false;
  const state = value as Record<string, unknown>;
  return (
    typeof state.title === "string" &&
    isRichTextValue(state.description) &&
    typeof state.teamId === "string" &&
    typeof state.idempotencyKey === "string" &&
    (state.statusId === null || typeof state.statusId === "string") &&
    (state.assigneeId === null || typeof state.assigneeId === "string") &&
    PRIORITIES.includes(state.priority as StoryPriority) &&
    Array.isArray(state.labelIds) &&
    state.labelIds.every((id) => typeof id === "string") &&
    (state.activeSheet === null ||
      ["team", "status", "priority", "assignee", "labels"].includes(
        state.activeSheet as string,
      ))
  );
};

export const formReducer = (
  state: FormState,
  action: FormAction,
): FormState => {
  switch (action.type) {
    case "setTitle":
      return { ...state, title: action.title };
    case "setDescription":
      return { ...state, description: action.description };
    case "setTeam":
      return {
        ...state,
        teamId: action.teamId,
        statusId: null,
        assigneeId: null,
        labelIds: [],
        activeSheet: null,
      };
    case "setStatus":
      return { ...state, statusId: action.statusId, activeSheet: null };
    case "setAssignee":
      return { ...state, assigneeId: action.assigneeId, activeSheet: null };
    case "setPriority":
      return { ...state, priority: action.priority, activeSheet: null };
    case "toggleLabel":
      return {
        ...state,
        labelIds: state.labelIds.includes(action.labelId)
          ? state.labelIds.filter((id) => id !== action.labelId)
          : [...state.labelIds, action.labelId],
      };
    case "setSheet":
      return { ...state, activeSheet: action.sheet };
  }
};
