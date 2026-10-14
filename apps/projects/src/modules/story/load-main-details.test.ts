/* global beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

const mockImportMainDetails = jest.fn();
const mockMainDetails = jest.fn();

jest.mock("./components/main-details", () => {
  mockImportMainDetails();
  return { MainDetails: mockMainDetails };
});

beforeEach(() => {
  jest.resetModules();
  mockImportMainDetails.mockReset();
});

it("shares an in-flight preload with rendering and subsequent intent", async () => {
  const { loadMainDetails, preloadMainDetails } = await import(
    "./load-main-details"
  );
  const preload = preloadMainDetails();
  const rendering = loadMainDetails();

  expect(loadMainDetails()).toBe(rendering);
  await preload;
  await expect(rendering).resolves.toBe(mockMainDetails);
  expect(loadMainDetails()).toBe(rendering);
  await preloadMainDetails();
  expect(mockImportMainDetails).toHaveBeenCalledTimes(1);
});

it("allows rendering to retry a failed speculative preload", async () => {
  mockImportMainDetails.mockImplementationOnce(() => {
    throw new Error("Chunk download failed");
  });
  const { loadMainDetails, preloadMainDetails } = await import(
    "./load-main-details"
  );

  await expect(preloadMainDetails()).resolves.toBeUndefined();
  await expect(loadMainDetails()).resolves.toBe(mockMainDetails);
  expect(mockImportMainDetails).toHaveBeenCalledTimes(2);
});

it("allows new intent to retry and does not hide failures from rendering", async () => {
  const failure = new Error("Chunk download failed");
  mockImportMainDetails.mockImplementation(() => {
    throw failure;
  });
  const { loadMainDetails, preloadMainDetails } = await import(
    "./load-main-details"
  );

  await expect(preloadMainDetails()).resolves.toBeUndefined();
  await expect(loadMainDetails()).rejects.toBe(failure);
  mockImportMainDetails.mockReset();
  await preloadMainDetails();
  await expect(loadMainDetails()).resolves.toBe(mockMainDetails);
  expect(mockImportMainDetails).toHaveBeenCalledTimes(1);
});
