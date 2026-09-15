import assert from "node:assert/strict";
import test from "node:test";
import {
  filterPropertyOptions,
  memberDisplayName,
  performPropertyChange,
  togglePropertyId,
} from "./picker-utils";

test("search retains every matching option beyond the old eight-item limit", () => {
  const options = Array.from({ length: 25 }, (_, index) => ({
    label: `Sprint ${index}`,
    description: index === 24 ? "Current iteration" : undefined,
  }));
  assert.equal(filterPropertyOptions(options, " sprint ").length, 25);
  assert.deepEqual(filterPropertyOptions(options, "CURRENT"), [options[24]]);
  assert.equal(filterPropertyOptions(options, "does not exist").length, 0);
});

test("unnamed members retain searchable display names instead of throwing", () => {
  assert.equal(
    memberDisplayName({
      fullName: null,
      username: "",
      email: "person@example.com",
    }),
    "person@example.com",
  );
  assert.equal(
    memberDisplayName({ fullName: "  ", username: "  jess  " }),
    "jess",
  );
  assert.equal(memberDisplayName({}), "Unnamed member");
});

test("label removal preserves other selections and never mutates the supplied list", () => {
  const ids = ["label-a", "label-b"];
  assert.deepEqual(togglePropertyId(ids, "label-a"), ["label-b"]);
  assert.deepEqual(togglePropertyId(ids, "label-c"), [...ids, "label-c"]);
  assert.deepEqual(ids, ["label-a", "label-b"]);
});

test("clear actions retain null and dismiss only after the mutation resolves", async () => {
  let finish!: () => void;
  const pending = new Promise<void>((resolve) => {
    finish = resolve;
  });
  let payload: string | null | undefined;
  let dismissed = false;
  let settled = false;
  const change = (id: string | null) => {
    payload = id;
    return pending;
  };
  const result = performPropertyChange(() => change(null), {
    onSuccess: () => {
      dismissed = true;
    },
    onError: () => assert.fail("Successful clear should not fail"),
    onSettled: () => {
      settled = true;
    },
  });
  await Promise.resolve();
  assert.equal(payload, null);
  assert.equal(dismissed, false);
  assert.equal(settled, false);
  finish();
  await result;
  assert.equal(dismissed, true);
  assert.equal(settled, true);
});

test("failed selections stay open, surface the server error, and release pending state", async () => {
  let error = "";
  let settled = false;
  await performPropertyChange(
    () => Promise.reject(new Error("This member is unavailable")),
    {
      onSuccess: () => assert.fail("Failed mutation must not dismiss"),
      onError: (message) => {
        error = message;
      },
      onSettled: () => {
        settled = true;
      },
    },
  );
  assert.equal(error, "This member is unavailable");
  assert.equal(settled, true);
});
