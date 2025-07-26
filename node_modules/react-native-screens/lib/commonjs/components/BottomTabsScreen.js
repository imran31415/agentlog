"use strict";
'use client';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;
var _react = _interopRequireDefault(require("react"));
var _BottomTabsScreenNativeComponent = _interopRequireDefault(require("../fabric/BottomTabsScreenNativeComponent"));
var _reactNative = require("react-native");
var _reactFreeze = require("react-freeze");
var _core = require("../core");
var _flags = require("../flags");
function _interopRequireDefault(e) { return e && e.__esModule ? e : { default: e }; }
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
// iOS-specific: SFSymbol usage

// iOS-specific: image as a template usage

// iOS-specific: SFSymbol, image as a template usage

/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
function BottomTabsScreen(props) {
  const componentNodeRef = _react.default.useRef(null);
  const componentNodeHandle = _react.default.useRef(-1);
  _react.default.useEffect(() => {
    if (componentNodeRef.current != null) {
      componentNodeHandle.current = (0, _reactNative.findNodeHandle)(componentNodeRef.current) ?? -1;
    } else {
      componentNodeHandle.current = -1;
    }
  }, []);
  const [nativeViewIsVisible, setNativeViewIsVisible] = _react.default.useState(false);
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
  let shouldFreeze = (0, _core.freezeEnabled)();
  if (_flags.featureFlags.experiment.controlledBottomTabs) {
    // If the tabs are JS controlled, we want to freeze only when given view is not focused && it is not currently visible
    shouldFreeze = shouldFreeze && !nativeViewIsVisible && !isFocused;
  } else {
    shouldFreeze = shouldFreeze && !nativeViewIsVisible;
  }
  const onWillAppearCallback = _react.default.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onWillAppear received`);
    setNativeViewIsVisible(true);
    onWillAppear?.(event);
  }, [onWillAppear]);
  const onDidAppearCallback = _react.default.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onDidAppear received`);
    onDidAppear?.(event);
  }, [onDidAppear]);
  const onWillDisappearCallback = _react.default.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onWillDisappear received`);
    onWillDisappear?.(event);
  }, [onWillDisappear]);
  const onDidDisappearCallback = _react.default.useCallback(event => {
    console.log(`TabsScreen [${componentNodeHandle.current}] onDidDisappear received`);
    setNativeViewIsVisible(false);
    onDidDisappear?.(event);
  }, [onDidDisappear]);
  console.info(`TabsScreen [${componentNodeHandle.current ?? -1}] render; tabKey: ${rest.tabKey} shouldFreeze: ${shouldFreeze}, isFocused: ${isFocused} nativeViewIsVisible: ${nativeViewIsVisible}`);
  const iconProps = parseIconsToNativeProps(icon, selectedIcon);
  return /*#__PURE__*/_react.default.createElement(_BottomTabsScreenNativeComponent.default, _extends({
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
  }, rest), /*#__PURE__*/_react.default.createElement(_reactFreeze.Freeze, {
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
var _default = exports.default = BottomTabsScreen;
const styles = _reactNative.StyleSheet.create({
  fillParent: {
    position: 'absolute',
    flex: 1,
    width: '100%',
    height: '100%'
  }
});
//# sourceMappingURL=BottomTabsScreen.js.map