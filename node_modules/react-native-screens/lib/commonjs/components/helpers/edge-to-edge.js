"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.EDGE_TO_EDGE = void 0;
exports.transformEdgeToEdgeProps = transformEdgeToEdgeProps;
var _reactNativeIsEdgeToEdge = require("react-native-is-edge-to-edge");
const EDGE_TO_EDGE = exports.EDGE_TO_EDGE = (0, _reactNativeIsEdgeToEdge.isEdgeToEdge)();
function transformEdgeToEdgeProps(props) {
  const {
    // Filter out edge-to-edge related props
    statusBarColor,
    statusBarTranslucent,
    navigationBarColor,
    navigationBarTranslucent,
    ...rest
  } = props;
  if (__DEV__) {
    (0, _reactNativeIsEdgeToEdge.controlEdgeToEdgeValues)({
      statusBarColor,
      statusBarTranslucent,
      navigationBarColor,
      navigationBarTranslucent
    });
  }
  return rest;
}
//# sourceMappingURL=edge-to-edge.js.map