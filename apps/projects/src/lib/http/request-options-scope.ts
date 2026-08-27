export type RequestOptionsScope = Readonly<{
  signal: AbortSignal;
}>;

type RequestOptionsScopeResolver = () => RequestOptionsScope | undefined;

const REQUEST_OPTIONS_SCOPE_RESOLVER_KEY = Symbol.for(
  "fortyone.http-request-options-scope-resolver",
);

type GlobalWithRequestOptionsScopeResolver = typeof globalThis & {
  [REQUEST_OPTIONS_SCOPE_RESOLVER_KEY]?: RequestOptionsScopeResolver;
};

const getGlobalState = () =>
  globalThis as GlobalWithRequestOptionsScopeResolver;

export const getRequestOptionsScope = () =>
  getGlobalState()[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY]?.();

export const installRequestOptionsScopeResolver = (
  resolver: RequestOptionsScopeResolver,
) => {
  const globalState = getGlobalState();
  const previousResolver = globalState[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY];
  globalState[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY] = resolver;

  return () => {
    if (globalState[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY] !== resolver) return;

    if (previousResolver) {
      globalState[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY] = previousResolver;
      return;
    }

    globalState[REQUEST_OPTIONS_SCOPE_RESOLVER_KEY] = undefined;
  };
};
