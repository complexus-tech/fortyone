import { ApiError } from "@/lib/http";
import { mayaSkillsKey, useMayaSkillMutations } from "./hooks";

const mockInvalidate = jest.fn();
const mockMutationOptions: { onError?: (error: Error) => Promise<void> }[] = [];
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "owner" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@tanstack/react-query", () => ({
  useMutation: (options: { onError?: (error: Error) => Promise<void> }) => {
    mockMutationOptions.push(options);
    return {};
  },
  useQueryClient: () => ({ invalidateQueries: mockInvalidate }),
  useQuery: jest.fn(),
}));

beforeEach(() => {
  mockMutationOptions.length = 0;
  mockInvalidate.mockReset().mockResolvedValue(undefined);
});

it("isolates personal skill caches by both workspace and user", () => {
  expect(mayaSkillsKey("acme", "owner")).not.toEqual(
    mayaSkillsKey("acme", "another-user"),
  );
  expect(mayaSkillsKey("acme", "owner")).not.toEqual(
    mayaSkillsKey("another-workspace", "owner"),
  );
});

it("refreshes the owned list after a conflict before allowing an edit retry", async () => {
  useMayaSkillMutations();
  const onError = mockMutationOptions[1].onError!;
  await onError(new ApiError("Skill changed elsewhere", 409, null));
  expect(mockInvalidate).toHaveBeenCalledWith({
    queryKey: mayaSkillsKey("acme", "owner"),
  });
  mockInvalidate.mockClear();
  await onError(new Error("Connection lost"));
  expect(mockInvalidate).not.toHaveBeenCalled();
});
