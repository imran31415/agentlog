"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;
var _react = _interopRequireDefault(require("react"));
var _reactNative = require("react-native");
var _ScreenStackHostNativeComponent = _interopRequireDefault(require("../../fabric/gamma/ScreenStackHostNativeComponent"));
function _interopRequireDefault(e) { return e && e.__esModule ? e : { default: e }; }
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function ScreenStackHost({
  children
}) {
  return /*#__PURE__*/_react.default.createElement(_ScreenStackHostNativeComponent.default, {
    style: styles.container
  }, children);
}
const styles = _reactNative.StyleSheet.create({
  container: {
    flex: 1
  }
});
var _default = exports.default = ScreenStackHost;
//# sourceMappingURL=ScreenStackHost.js.map