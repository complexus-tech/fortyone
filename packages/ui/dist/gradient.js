"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Gradient = Gradient;
const jsx_runtime_1 = require("react/jsx-runtime");
function Gradient({ conic, className, small, }) {
    return ((0, jsx_runtime_1.jsx)("span", { className: `ui:absolute ui:mix-blend-normal ui:will-change-[filter] ui:rounded-[100%] ${small ? "ui:blur-[32px]" : "ui:blur-[75px]"} ${conic
            ? "ui:bg-gradient-to-r ui:bg-red-1000 ui:from-10% ui:via-purple-1000 ui:via-30% ui:to-blue-1000 ui:to-100%"
            : ""} ${className ?? ""}` }));
}
