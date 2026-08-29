const React = require("react");

const createPrimitive = (displayName) => {
  const Primitive = React.forwardRef(({ children, ...props }, ref) =>
    React.createElement("div", { ...props, ref }, children),
  );
  Primitive.displayName = displayName;
  return Primitive;
};

module.exports = {
  Panel: createPrimitive("Panel"),
  PanelGroup: createPrimitive("PanelGroup"),
  PanelResizeHandle: createPrimitive("PanelResizeHandle"),
};
