import React from 'react';
import type { ViewProps } from 'react-native';
import type { NativeProps } from '../../fabric/gamma/ScreenStackHostNativeComponent';
export type ScreenStackNativeProps = NativeProps & {};
type ScreenStackHostProps = {
    children?: ViewProps['children'];
} & ScreenStackNativeProps;
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
declare function ScreenStackHost({ children }: ScreenStackHostProps): React.JSX.Element;
export default ScreenStackHost;
//# sourceMappingURL=ScreenStackHost.d.ts.map