"use strict";
'use client';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;
var _react = _interopRequireDefault(require("react"));
var _BottomTabsNativeComponent = _interopRequireDefault(require("../fabric/BottomTabsNativeComponent"));
var _reactNative = require("react-native");
var _flags = _interopRequireDefault(require("../flags"));
function _interopRequireDefault(e) { return e && e.__esModule ? e : { default: e }; }
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function BottomTabs(props) {
  console.info(`BottomTabs render`);
  const {
    onNativeFocusChange,
    experimentalControlNavigationStateInJS = _flags.default.experiment.controlledBottomTabs,
    ...filteredProps
  } = props;
  const componentNodeRef = _react.default.useRef(null);
  const componentNodeHandle = _react.default.useRef(-1);
  _react.default.useEffect(() => {
    if (componentNodeRef.current != null) {
      componentNodeHandle.current = (0, _reactNative.findNodeHandle)(componentNodeRef.current) ?? -1;
    } else {
      componentNodeHandle.current = -1;
    }
  }, []);
  const onNativeFocusChangeCallback = _react.default.useCallback(event => {
    console.log(`BottomTabs [${componentNodeHandle.current ?? -1}] onNativeFocusChange: ${JSON.stringify(event.nativeEvent)}`);
    onNativeFocusChange?.(event);
  }, [onNativeFocusChange]);
  return /*#__PURE__*/_react.default.createElement(_BottomTabsNativeComponent.default, _extends({
    style: styles.fillParent,
    onNativeFocusChange: onNativeFocusChangeCallback,
    controlNavigationStateInJS: experimentalControlNavigationStateInJS
    // @ts-ignore suppress ref - debug only
    ,
    ref: componentNodeRef
  }, filteredProps), filteredProps.children);
}
var _default = exports.default = BottomTabs;
const styles = _reactNative.StyleSheet.create({
  fillParent: {
    flex: 1,
    width: '100%',
    height: '100%'
  }
});
//# sourceMappingURL=BottomTabs.js.map