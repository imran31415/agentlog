"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = exports.StackScreenLifecycleState = void 0;
var _react = _interopRequireDefault(require("react"));
var _reactNative = require("react-native");
var _StackScreenNativeComponent = _interopRequireDefault(require("../../fabric/gamma/StackScreenNativeComponent"));
function _interopRequireDefault(e) { return e && e.__esModule ? e : { default: e }; }
const StackScreenLifecycleState = exports.StackScreenLifecycleState = {
  INITIAL: 0,
  DETACHED: 1,
  ATTACHED: 2
};
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function StackScreen({
  children,
  // Control
  maxLifecycleState,
  screenKey,
  // Events
  onWillAppear,
  onWillDisappear,
  onDidAppear,
  onDidDisappear,
  // Custom events
  onPop
}) {
  const handleOnDidDisappear = _react.default.useCallback(e => {
    onDidDisappear?.(e);
    onPop?.(screenKey);
  }, [onDidDisappear, onPop, screenKey]);
  return /*#__PURE__*/_react.default.createElement(_StackScreenNativeComponent.default, {
    style: _reactNative.StyleSheet.absoluteFill
    // Control
    ,
    maxLifecycleState: maxLifecycleState,
    screenKey: screenKey
    // Events
    ,
    onWillAppear: onWillAppear,
    onDidAppear: onDidAppear,
    onWillDisappear: onWillDisappear,
    onDidDisappear: handleOnDidDisappear
  }, children);
}
var _default = exports.default = StackScreen;
//# sourceMappingURL=StackScreen.js.map