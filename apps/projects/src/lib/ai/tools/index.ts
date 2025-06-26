import { navigationTool } from "./navigation";
import { storiesTool } from "./stories";
import { sprintsTool } from "./sprints";

export { navigationTool } from "./navigation";
export { storiesTool } from "./stories";
export { sprintsTool } from "./sprints";

// Export all tools as an array for easy registration
export const allTools = [navigationTool, storiesTool, sprintsTool];
