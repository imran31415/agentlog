'use client';

function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
import React from 'react';
import BottomTabsNativeComponent from '../fabric/BottomTabsNativeComponent';
import { findNodeHandle, StyleSheet } from 'react-native';
import featureFlags from '../flags';
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function BottomTabs(props) {
  console.info(`BottomTabs render`);
  const {
    onNativeFocusChange,
    experimentalControlNavigationStateInJS = featureFlags.experiment.controlledBottomTabs,
    ...filteredProps
  } = props;
  const componentNodeRef = React.useRef(null);
  const componentNodeHandle = React.useRef(-1);
  React.useEffect(() => {
    if (componentNodeRef.current != null) {
      componentNodeHandle.current = findNodeHandle(componentNodeRef.current) ?? -1;
    } else {
      componentNodeHandle.current = -1;
    }
  }, []);
  const onNativeFocusChangeCallback = React.useCallback(event => {
    console.log(`BottomTabs [${componentNodeHandle.current ?? -1}] onNativeFocusChange: ${JSON.stringify(event.nativeEvent)}`);
    onNativeFocusChange?.(event);
  }, [onNativeFocusChange]);
  return /*#__PURE__*/React.createElement(BottomTabsNativeComponent, _extends({
    style: styles.fillParent,
    onNativeFocusChange: onNativeFocusChangeCallback,
    controlNavigationStateInJS: experimentalControlNavigationStateInJS
    // @ts-ignore suppress ref - debug only
    ,
    ref: componentNodeRef
  }, filteredProps), filteredProps.children);
}
export default BottomTabs;
const styles = StyleSheet.create({
  fillParent: {
    flex: 1,
    width: '100%',
    height: '100%'
  }
});
//# sourceMappingURL=BottomTabs.js.map