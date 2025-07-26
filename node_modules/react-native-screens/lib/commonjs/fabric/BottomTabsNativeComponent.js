"use strict";
'use client';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;
var _codegenNativeComponent = _interopRequireDefault(require("react-native/Libraries/Utilities/codegenNativeComponent"));
function _interopRequireDefault(e) { return e && e.__esModule ? e : { default: e }; }
// TODO: Report issue on RN repo, that nesting color value inside a struct does not work.
// Generated code is ok, but the value is not passed down correctly - whatever color is set
// host component receives RGBA(0, 0, 0, 0) anyway.
// type TabBarAppearance = {
//   backgroundColor?: ColorValue;
// };
var _default = exports.default = (0, _codegenNativeComponent.default)('RNSBottomTabs', {
  interfaceOnly: true
});
//# sourceMappingURL=BottomTabsNativeComponent.js.map