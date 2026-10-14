import { createRef } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { StrategyNodePositions } from "./strategy-map-layout";
import {
  useStrategyMapCanvasInteractions,
  useStrategyMapCanvasInteractionState,
} from "./use-strategy-map-canvas-interactions";

const NODE_ID = "key-result:activation";

const renderInteractions = () => {
  const viewportRef = createRef<HTMLDivElement>();
  const positions: StrategyNodePositions = {
    [NODE_ID]: { x: 100, y: 100 },
  };
  const positionsRef = { current: positions };
  const persistPositions = jest.fn();
  const setTransientPositions = jest.fn();
  const onSelect = jest.fn();
  const Harness = () => {
    const state = useStrategyMapCanvasInteractionState();
    const handlers = useStrategyMapCanvasInteractions({
      dimensionsRef: { current: {} },
      expandedObjectiveIds: new Set(),
      keyResultObjectiveIdByNodeId: new Map([[NODE_ID, "objective-1"]]),
      keyResultsByObjective: new Map(),
      layout: { height: 2000, positions, width: 2000 },
      onAlign: jest.fn(),
      persistPositions,
      pillarByObjectiveId: new Map(),
      pillars: [],
      positionOverrides: {},
      positions,
      positionsRef,
      setTransientPositions,
      state,
      viewportRef,
      zoom: 1,
    });

    return (
      <>
        <div
          data-testid="viewport"
          onPointerDown={handlers.handleViewportPointerDown}
          onPointerMove={handlers.handleViewportPointerMove}
          onPointerUp={handlers.handleViewportPointerUp}
          ref={viewportRef}
        >
          <div
            data-node-id={NODE_ID}
            data-testid="card"
            onPointerDown={(event) => {
              handlers.handleNodePointerDown(NODE_ID, event);
            }}
            onPointerMove={handlers.handleNodePointerMove}
            onPointerUp={(event) => {
              handlers.finishNodeDrag(event, true, onSelect);
            }}
          >
            Improve activation
          </div>
        </div>
        <output data-testid="dragging">{state.draggingNodeId ?? "none"}</output>
        <output data-testid="panning">{String(state.isPanning)}</output>
      </>
    );
  };

  render(<Harness />);
  const card = screen.getByTestId("card");
  const viewport = screen.getByTestId("viewport");
  const capturePointer = jest.fn();
  const releasePointer = jest.fn();
  for (const element of [card, viewport]) {
    Object.defineProperties(element, {
      hasPointerCapture: { value: () => true },
      releasePointerCapture: { value: releasePointer },
      setPointerCapture: { value: capturePointer },
    });
  }
  viewport.scrollLeft = 300;
  viewport.scrollTop = 300;

  return {
    capturePointer,
    card,
    onSelect,
    persistPositions,
    positionsRef,
    releasePointer,
    setTransientPositions,
    viewport,
  };
};

const dispatchPointer = (
  element: HTMLElement,
  type: "pointerdown" | "pointermove" | "pointerup",
  options: MouseEventInit = {},
) => {
  const event = new PointerEvent(type, {
    bubbles: true,
    cancelable: true,
    clientX: 100,
    clientY: 100,
    ...options,
  });
  Object.defineProperty(event, "pointerId", { value: 1 });
  fireEvent(element, event);
  return event;
};

describe("strategy map context-menu gestures", () => {
  it.each([
    { label: "secondary mouse button", button: 2, ctrlKey: false },
    { label: "Control-primary click", button: 0, ctrlKey: true },
  ])("leaves $label on cards available to the context menu", (options) => {
    const {
      capturePointer,
      card,
      onSelect,
      persistPositions,
      positionsRef,
      setTransientPositions,
    } = renderInteractions();

    const down = dispatchPointer(card, "pointerdown", options);
    dispatchPointer(card, "pointermove", { clientX: 180, clientY: 170 });
    dispatchPointer(card, "pointerup", { clientX: 180, clientY: 170 });

    expect(down.defaultPrevented).toBe(false);
    expect(capturePointer).not.toHaveBeenCalled();
    expect(screen.getByTestId("dragging")).toHaveTextContent("none");
    expect(screen.getByTestId("panning")).toHaveTextContent("false");
    expect(positionsRef.current[NODE_ID]).toEqual({ x: 100, y: 100 });
    expect(setTransientPositions).not.toHaveBeenCalled();
    expect(persistPositions).not.toHaveBeenCalled();
    expect(onSelect).not.toHaveBeenCalled();
    expect(document.body.style.cursor).toBe("");
  });

  it.each([
    { label: "secondary mouse button", button: 2, ctrlKey: false },
    { label: "Control-primary click", button: 0, ctrlKey: true },
  ])("does not pan the canvas on $label", (options) => {
    const { capturePointer, viewport } = renderInteractions();
    const down = dispatchPointer(viewport, "pointerdown", options);
    dispatchPointer(viewport, "pointermove", { clientX: 180, clientY: 170 });
    dispatchPointer(viewport, "pointerup", { clientX: 180, clientY: 170 });

    expect(down.defaultPrevented).toBe(false);
    expect(capturePointer).not.toHaveBeenCalled();
    expect(screen.getByTestId("panning")).toHaveTextContent("false");
    expect(viewport.scrollLeft).toBe(300);
    expect(viewport.scrollTop).toBe(300);
  });

  it("still drags cards with the ordinary primary button", () => {
    const {
      capturePointer,
      card,
      onSelect,
      persistPositions,
      positionsRef,
      releasePointer,
      viewport,
    } = renderInteractions();
    const down = dispatchPointer(card, "pointerdown", { button: 0 });
    expect(down.defaultPrevented).toBe(true);
    expect(capturePointer).toHaveBeenCalledWith(1);
    expect(screen.getByTestId("dragging")).toHaveTextContent(NODE_ID);
    dispatchPointer(card, "pointermove", { clientX: 180, clientY: 170 });
    dispatchPointer(card, "pointerup", { clientX: 180, clientY: 170 });

    expect(positionsRef.current[NODE_ID]).toEqual({ x: 180, y: 170 });
    expect(releasePointer).toHaveBeenCalledWith(1);
    expect(persistPositions).toHaveBeenCalledTimes(1);
    expect(onSelect).not.toHaveBeenCalled();
    expect(screen.getByTestId("dragging")).toHaveTextContent("none");
    expect(viewport.scrollLeft).toBe(300);
  });

  it("still pans the canvas with the ordinary primary button", () => {
    const { capturePointer, releasePointer, viewport } = renderInteractions();
    const down = dispatchPointer(viewport, "pointerdown", { button: 0 });
    expect(down.defaultPrevented).toBe(true);
    expect(capturePointer).toHaveBeenCalledWith(1);
    expect(screen.getByTestId("panning")).toHaveTextContent("true");

    dispatchPointer(viewport, "pointermove", { clientX: 180, clientY: 170 });
    expect(viewport.scrollLeft).toBe(220);
    expect(viewport.scrollTop).toBe(230);
    dispatchPointer(viewport, "pointerup", { clientX: 180, clientY: 170 });
    expect(releasePointer).toHaveBeenCalledWith(1);
    expect(screen.getByTestId("panning")).toHaveTextContent("false");
  });

  it("still selects cards on a primary click without movement", () => {
    const { card, onSelect } = renderInteractions();
    dispatchPointer(card, "pointerdown", { button: 0 });
    dispatchPointer(card, "pointerup", { button: 0 });
    expect(onSelect).toHaveBeenCalledTimes(1);
  });
});
