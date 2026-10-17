import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { chatCapabilityKeys, workspaceKeys } from "@/constants/keys";
import { updateWorkspaceSettingsAction } from "@/lib/actions/workspaces/update";
import { get } from "@/lib/http";
import type { WorkspaceSettings } from "@/types";
import { useChatCapabilities } from "./chat-capabilities";
import { useUpdateWorkspaceSettingsMutation } from "./update-settings";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "viewer-one" } } }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/actions/workspaces/update", () => ({
  updateWorkspaceSettingsAction: jest.fn(),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const settings: WorkspaceSettings = {
  storyTerm: "task",
  sprintTerm: "sprint",
  objectiveTerm: "objective",
  keyResultTerm: "key result",
  objectiveEnabled: true,
  keyResultEnabled: true,
  chatEnabled: true,
  workingDays: [1, 2, 3, 4, 5],
  workingStartMinute: 540,
  workingEndMinute: 1020,
};
let client: QueryClient;
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  jest.clearAllMocks();
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  client.setQueryData(workspaceKeys.settings("acme"), settings);
  jest.mocked(get).mockResolvedValue({
    data: { enabled: true, attachmentsEnabled: true },
  });
});
afterEach(() => {
  cleanup();
  client.clear();
});

it("updates the existing settings endpoint and immediately rechecks capabilities", async () => {
  const view = renderHook(
    () => ({
      chat: useChatCapabilities(),
      update: useUpdateWorkspaceSettingsMutation(),
    }),
    { wrapper: Wrapper },
  );
  await waitFor(() => {
    expect(view.result.current.chat.enabled).toBe(true);
  });
  jest.mocked(get).mockResolvedValue({
    data: { enabled: false, attachmentsEnabled: false },
  });
  jest.mocked(updateWorkspaceSettingsAction).mockResolvedValue({
    data: { ...settings, chatEnabled: false },
  });
  await act(async () => {
    await view.result.current.update.mutateAsync({ chatEnabled: false });
  });
  await waitFor(() => {
    expect(view.result.current.chat.enabled).toBe(false);
  });
  expect(view.result.current.chat.attachmentsEnabled).toBe(false);
  expect(updateWorkspaceSettingsAction).toHaveBeenCalledWith(
    { chatEnabled: false },
    "acme",
  );
  expect(get).toHaveBeenCalledTimes(2);
  expect(client.getQueryData(workspaceKeys.settings("acme"))).toEqual({
    ...settings,
    chatEnabled: false,
  });
});

it("restores optimistic settings after a rejected save and rechecks the server", async () => {
  jest.mocked(updateWorkspaceSettingsAction).mockResolvedValue({
    error: { message: "Only workspace admins can change settings" },
  });
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const view = renderHook(useUpdateWorkspaceSettingsMutation, {
    wrapper: Wrapper,
  });
  await act(async () => {
    await expect(
      view.result.current.mutateAsync({ chatEnabled: false }),
    ).rejects.toThrow("Only workspace admins can change settings");
  });
  expect(client.getQueryData(workspaceKeys.settings("acme"))).toEqual(settings);
  expect(toast.error).toHaveBeenCalledWith(
    "Failed to update workspace settings",
    expect.objectContaining({
      description: "Only workspace admins can change settings",
    }),
  );
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: chatCapabilityKeys.workspace("acme"),
  });
});

it("restores attachment and recording controls when Chat is enabled", async () => {
  jest.mocked(get).mockResolvedValue({
    data: { enabled: false, attachmentsEnabled: false },
  });
  const view = renderHook(
    () => ({
      chat: useChatCapabilities(),
      update: useUpdateWorkspaceSettingsMutation(),
    }),
    { wrapper: Wrapper },
  );
  await waitFor(() => {
    expect(view.result.current.chat.isSuccess).toBe(true);
  });
  jest.mocked(get).mockResolvedValue({
    data: { enabled: true, attachmentsEnabled: true },
  });
  jest
    .mocked(updateWorkspaceSettingsAction)
    .mockResolvedValue({ data: settings });
  await act(async () => {
    await view.result.current.update.mutateAsync({ chatEnabled: true });
  });
  await waitFor(() => {
    expect(view.result.current.chat.enabled).toBe(true);
  });
  expect(view.result.current.chat.attachmentsEnabled).toBe(true);
});

it("does not recheck Chat for unrelated workspace settings", async () => {
  jest.mocked(updateWorkspaceSettingsAction).mockResolvedValue({
    data: { ...settings, objectiveEnabled: false },
  });
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const view = renderHook(useUpdateWorkspaceSettingsMutation, {
    wrapper: Wrapper,
  });
  await act(async () => {
    await view.result.current.mutateAsync({ objectiveEnabled: false });
  });
  expect(invalidate).toHaveBeenCalledTimes(1);
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: workspaceKeys.settings("acme"),
  });
});
