function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
import React from 'react';
import { StyleSheet } from 'react-native';
import SplitViewHostNativeComponent from '../../fabric/gamma/SplitViewHostNativeComponent';
// According to the UIKit documentation: https://developer.apple.com/documentation/uikit/uisplitviewcontroller/displaymode-swift.enum
// Only specific pairs for displayMode - splitBehavior are valid and others may lead to unexpected results.
// Therefore, we're adding check on the JS side to return a feedback to the client when that pairing isn't valid.
// However, we're not blocking these props to be set on the native side, because it doesn't crash, just the result or transitions may not work as expected.
const displayModeForSplitViewCompatibilityMap = {
  tile: ['secondaryOnly', 'oneBesideSecondary', 'twoBesideSecondary'],
  overlay: ['secondaryOnly', 'oneOverSecondary', 'twoOverSecondary'],
  displace: ['secondaryOnly', 'oneBesideSecondary', 'twoDisplaceSecondary'],
  automatic: [] // placeholder for satisfying types; we'll handle it specially in logic
};
const isValidDisplayModeForSplitBehavior = (displayMode, splitBehavior) => {
  if (splitBehavior === 'automatic') {
    // for automatic we cannot easily verify the compatibility, because it depends on the system preference for display mode, therefore we're assuming that 'automatic' has only valid combinations
    return true;
  }
  return displayModeForSplitViewCompatibilityMap[splitBehavior].includes(displayMode);
};

/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function SplitViewHost(props) {
  const {
    displayMode,
    splitBehavior
  } = props;
  React.useEffect(() => {
    if (displayMode && splitBehavior) {
      const isValid = isValidDisplayModeForSplitBehavior(displayMode, splitBehavior);
      if (!isValid) {
        const validDisplayModes = displayModeForSplitViewCompatibilityMap[splitBehavior];
        console.warn(`Invalid display mode "${displayMode}" for split behavior "${splitBehavior}".` + `\nValid modes for "${splitBehavior}" are: ${validDisplayModes.join(', ')}.`);
      }
    }
  }, [displayMode, splitBehavior]);
  return /*#__PURE__*/React.createElement(SplitViewHostNativeComponent, _extends({}, props, {
    style: styles.container
  }), props.children);
}
const styles = StyleSheet.create({
  container: {
    flex: 1
  }
});
export default SplitViewHost;
//# sourceMappingURL=SplitViewHost.js.map