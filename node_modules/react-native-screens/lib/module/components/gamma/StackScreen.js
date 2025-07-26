import React from 'react';
import { StyleSheet } from 'react-native';
import StackScreenNativeComponent from '../../fabric/gamma/StackScreenNativeComponent';
export const StackScreenLifecycleState = {
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
  const handleOnDidDisappear = React.useCallback(e => {
    onDidDisappear?.(e);
    onPop?.(screenKey);
  }, [onDidDisappear, onPop, screenKey]);
  return /*#__PURE__*/React.createElement(StackScreenNativeComponent, {
    style: StyleSheet.absoluteFill
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
export default StackScreen;
//# sourceMappingURL=StackScreen.js.map