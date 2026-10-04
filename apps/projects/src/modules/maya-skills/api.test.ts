import { get, post, put, remove } from "@/lib/http";
import {
  createMayaSkill,
  deleteMayaSkill,
  listMayaSkills,
  updateMayaSkill,
} from "./api";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  remove: jest.fn(),
}));

const ctx = { session: { token: "session" }, workspaceSlug: "acme" };
const input = {
  name: " Weekly update ",
  description: " Team summary ",
  instructions: " Review current work. ",
};

beforeEach(() => {
  jest.clearAllMocks();
});

it("requires authenticated workspace context before making a request", () => {
  const anonymous = { workspaceSlug: "acme" };
  expect(() => listMayaSkills(anonymous)).toThrow("Sign in");
  expect(() => createMayaSkill(input, anonymous)).toThrow("Sign in");
  expect(() =>
    updateMayaSkill("skill", { ...input, updatedAt: "version" }, anonymous),
  ).toThrow("Sign in");
  expect(() => deleteMayaSkill("skill", anonymous)).toThrow("Sign in");
  expect(get).not.toHaveBeenCalled();
  expect(post).not.toHaveBeenCalled();
  expect(put).not.toHaveBeenCalled();
  expect(remove).not.toHaveBeenCalled();
});

it("sends trimmed instructions without client ownership fields and retains the original edit version", () => {
  createMayaSkill(input, ctx);
  expect(post).toHaveBeenCalledWith(
    "maya/skills",
    {
      name: "Weekly update",
      description: "Team summary",
      instructions: "Review current work.",
    },
    ctx,
    undefined,
    expect.any(Function),
  );
  updateMayaSkill(
    "skill",
    { ...input, updatedAt: "2026-10-04T09:00:00Z" },
    ctx,
  );
  expect(put).toHaveBeenCalledWith(
    "maya/skills/skill",
    {
      name: "Weekly update",
      description: "Team summary",
      instructions: "Review current work.",
      updatedAt: "2026-10-04T09:00:00Z",
    },
    ctx,
    undefined,
    expect.any(Function),
  );
});

it("decodes the full owned list without silently truncating it", () => {
  listMayaSkills(ctx);
  const decoder = jest.mocked(get).mock.calls[0][3]!;
  const skills = Array.from({ length: 200 }, (_, index) => ({
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    name: `Skill ${index}`,
    description: "",
    instructions: "Review the work",
    createdAt: "2026-10-04T09:00:00Z",
    updatedAt: "2026-10-04T09:00:00Z",
  }));
  expect(decoder({ data: skills })).toHaveLength(200);
  expect(() => decoder({ data: [{ name: "Incomplete" }] })).toThrow();
});

it("decodes server-valid astral characters using Unicode character limits", () => {
  listMayaSkills(ctx);
  const decoder = jest.mocked(get).mock.calls[0][3]!;
  const skills = [41, 80].map((length) => ({
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    name: "😀".repeat(length),
    description: "😀".repeat(300),
    instructions: "😀".repeat(12_000),
    createdAt: "2026-10-04T09:00:00Z",
    updatedAt: "2026-10-04T09:00:00Z",
  }));
  expect(decoder({ data: skills })).toEqual(skills);
});

it.each([
  { field: "name", limit: 80, message: "Keep the name" },
  { field: "description", limit: 300, message: "Keep the description" },
  { field: "instructions", limit: 12_000, message: "Keep instructions" },
])(
  "rejects $field beyond its Unicode character limit before requesting a save",
  ({ field, limit, message }) => {
    expect(() =>
      createMayaSkill(
        { ...input, [field]: ` ${"😀".repeat(limit + 1)} ` },
        ctx,
      ),
    ).toThrow(message);
    expect(post).not.toHaveBeenCalled();
  },
);
