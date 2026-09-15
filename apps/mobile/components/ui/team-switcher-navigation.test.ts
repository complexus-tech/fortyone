import type { StackNavigationState } from "expo-router/react-navigation";
import assert from "node:assert/strict";
import test from "node:test";
// Exercise the installed router without loading React Native's native runtime.
import { StackRouter } from "expo-router/build/react-navigation/routers/StackRouter";
import { createTeamReplacementAction } from "./team-switcher-navigation";

const config = {
  routeNames: ["(tabs)", "teams/[teamId]"],
  routeParamList: {},
  routeGetIdList: {},
};

function initialState(): StackNavigationState<
  Record<string, object | undefined>
> {
  return {
    stale: false,
    type: "stack",
    key: "root-stack",
    index: 1,
    routeNames: config.routeNames,
    preloadedRoutes: [],
    routes: [
      { key: "home", name: "(tabs)" },
      {
        key: "team-A",
        name: "teams/[teamId]",
        params: { teamId: "A" },
        state: {
          key: "team-A-stack",
          index: 0,
          routes: [
            { key: "team-A-index", name: "index", params: { teamId: "A" } },
          ],
        },
      },
    ],
  };
}

test("team replacement updates the root route and its nested destination", () => {
  const router = StackRouter({});
  const before = initialState();
  const action = createTeamReplacementAction(before, "B");
  assert.ok(action);
  assert.equal(action.target, before.key);
  const after = router.getStateForAction(before, action, config);
  assert.ok(after);
  assert.equal(after.routes.length, 2);
  assert.equal(after.index, 1);
  assert.strictEqual(after.routes[0], before.routes[0]);
  assert.notEqual(after.routes[1].key, before.routes[1].key);
  assert.equal(after.routes[1].state, undefined);
  assert.deepEqual(after.routes[1].params, {
    teamId: "B",
    screen: "index",
    params: { teamId: "B" },
  });
});

test("successive team replacements keep Home as the back destination", () => {
  const router = StackRouter({});
  let state = initialState();
  for (const teamId of ["B", "C"]) {
    const action = createTeamReplacementAction(state, teamId);
    assert.ok(action);
    const next = router.getStateForAction(state, action, config);
    assert.ok(next);
    state = router.getRehydratedState(next, config);
  }
  const back = router.getStateForAction(state, { type: "GO_BACK" }, config);
  assert.ok(back);
  assert.equal(back.routes.length, 1);
  assert.equal(back.index, 0);
  assert.equal(back.routes[0].key, "home");
});

test("a delayed selection cannot replace a different root screen", () => {
  const state = initialState();
  assert.equal(createTeamReplacementAction(undefined, "B"), null);
  assert.equal(createTeamReplacementAction({ ...state, index: 0 }, "B"), null);
  assert.equal(
    createTeamReplacementAction(
      {
        ...state,
        routes: [...state.routes, { key: "settings", name: "settings" }],
        index: 2,
      },
      "B",
    ),
    null,
  );
});
