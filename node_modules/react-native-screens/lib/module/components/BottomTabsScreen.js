'use client';

function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
import React from 'react';
import BottomTabsScreenNativeComponent from '../fabric/BottomTabsScreenNativeComponent';
import { StyleSheet, findNodeHandle } from 'react-native';
import { Freeze } from 'react-freeze';
import { freezeEnabled } from '../core';
import { featureFlags } from '../flags';

// iOS-specific: SFSymbol usage

// iOS-specific: image as a template usage

// iOS-specific: SFSymbol, image as a template usage

/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function BottomTabsScreen(props) {
  const componentNodeRef = React.useRef(null);
  const componentNodeHandle = React.useRef(-1);
  React.useEffect(() => {
    if (componentNodeRef.current != null) {
      componentNodeHandle.current = findNodeHandle(componentNodeRef.current) ?? -1;
    } else {
      componentNodeHandle.current = -1;
    }
  }, []);
  const [nativeViewIsVisible, setNativeViewIsVisible] = React.useState(false);
  const {
    onWillAppear,
    onDidAppear,
    onWillDisappear,
    onDidDisappear,
    isFocused = false,
    icon,
    selectedIcon,
    ...rest
  } = props;
  let shouldFreeze = freezeEnabled();
  if (featureFlags.experiment.controlledBottomTabs) {
    // If the tabs are JS controlled, we want to freeze only when given view is not focused && it is not currently visible
    shouldFreeze = shouldFreeze && !nativeViewIsVisible && !isFocused;
  } else {
    shouldFreeze = shouldFreeze && !nativeViewIsVisible;
  }
  const onWillAppearCallback = React.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onWillAppear received`);
    setNativeViewIsVisible(true);
    onWillAppear?.(event);
  }, [onWillAppear]);
  const onDidAppearCallback = React.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onDidAppear received`);
    onDidAppear?.(event);
  }, [onDidAppear]);
  const onWillDisappearCallback = React.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onWillDisappear received`);
    onWillDisappear?.(event);
  }, [onWillDisappear]);
  const onDidDisappearCallback = React.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onDidDisappear received`);
    setNativeViewIsVisible(false);
    onDidDisappear?.(event);
  }, [onDidDisappear]);
  console.info(`TabsScreen [${componentNodeHandle.current ?? -1}] render; tabKey: ${rest.tabKey} shouldFreeze: ${shouldFreeze}, isFocused: ${isFocused} nativeViewIsVisible: ${nativeViewIsVisible}`);
  const iconProps = parseIconsToNativeProps(icon, selectedIcon);
  return /*#__PURE__*/React.createElement(BottomTabsScreenNativeComponent, _extends({
    collapsable: false,
    style: styles.fillParent,
    onWillAppear: onWillAppearCallback,
    onDidAppear: onDidAppearCallback,
    onWillDisappear: onWillDisappearCallback,
    onDidDisappear: onDidDisappearCallback,
    isFocused: isFocused
  }, iconProps, {
    // @ts-ignore - This is debug only anyway
    ref: componentNodeRef
  }, rest), /*#__PURE__*/React.createElement(Freeze, {
    freeze: shouldFreeze,
    placeholder: rest.placeholder
  }, rest.children));
}
function parseIconToNativeProps(icon) {
  if (!icon) {
    return {};
  }
  if ('sfSymbolName' in icon) {
    // iOS-specific: SFSymbol usage
    return {
      iconType: 'sfSymbol',
      iconSfSymbolName: icon.sfSymbolName
    };
  } else if ('imageSource' in icon) {
    return {
      iconType: 'image',
      iconImageSource: icon.imageSource
    };
  } else if ('templateSource' in icon) {
    // iOS-specifig: image as a template usage
    return {
      iconType: 'template',
      iconImageSource: icon.templateSource
    };
  } else {
    // iOS-specific: SFSymbol, image as a template usage
    throw new Error('[RNScreens] Incorrect icon format. You must provide sfSymbolName, imageSource or templateSource.');
  }
}
function parseIconsToNativeProps(icon, selectedIcon) {
  const {
    iconImageSource,
    iconSfSymbolName,
    iconType
  } = parseIconToNativeProps(icon);
  const {
    iconImageSource: selectedIconImageSource,
    iconSfSymbolName: selectedIconSfSymbolName,
    iconType: selectedIconType
  } = parseIconToNativeProps(selectedIcon);
  if (iconType !== undefined && selectedIconType !== undefined && iconType !== selectedIconType) {
    throw new Error('[RNScreens] icon and selectedIcon must be same type.');
  } else if (iconType === undefined && selectedIconType !== undefined) {
    // iOS-specific: UIKit requirement
    throw new Error('[RNScreens] To use selectedIcon prop, the icon prop must also be provided.');
  }
  return {
    iconType,
    iconImageSource,
    iconSfSymbolName,
    selectedIconImageSource,
    selectedIconSfSymbolName
  };
}
export default BottomTabsScreen;
const styles = StyleSheet.create({
  fillParent: {
    position: 'absolute',
    flex: 1,
    width: '100%',
    height: '100%'
  }
});
//# sourceMappingURL=BottomTabsScreen.js.map