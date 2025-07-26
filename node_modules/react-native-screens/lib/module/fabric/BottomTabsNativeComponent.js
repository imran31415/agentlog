'use client';

import codegenNativeComponent from 'react-native/Libraries/Utilities/codegenNativeComponent';

// TODO: Report issue on RN repo, that nesting color value inside a struct does not work.
// Generated code is ok, but the value is not passed down correctly - whatever color is set
// host component receives RGBA(0, 0, 0, 0) anyway.
// type TabBarAppearance = {
//   backgroundColor?: ColorValue;
// };

export default codegenNativeComponent('RNSBottomTabs', {
  interfaceOnly: true
});
//# sourceMappingURL=BottomTabsNativeComponent.js.map