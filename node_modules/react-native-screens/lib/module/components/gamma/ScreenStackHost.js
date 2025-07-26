import React from 'react';
import { StyleSheet } from 'react-native';
import ScreenStackHostNativeComponent from '../../fabric/gamma/ScreenStackHostNativeComponent';
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function ScreenStackHost({
  children
}) {
  return /*#__PURE__*/React.createElement(ScreenStackHostNativeComponent, {
    style: styles.container
  }, children);
}
const styles = StyleSheet.create({
  container: {
    flex: 1
  }
});
export default ScreenStackHost;
//# sourceMappingURL=ScreenStackHost.js.map