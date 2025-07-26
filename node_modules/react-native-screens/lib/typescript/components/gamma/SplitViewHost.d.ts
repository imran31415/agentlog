import React from 'react';
import type { ViewProps } from 'react-native';
import type { NativeProps } from '../../fabric/gamma/SplitViewHostNativeComponent';
export type SplitViewNativeProps = NativeProps & {};
type SplitViewHostProps = {
    children?: ViewProps['children'];
} & SplitViewNativeProps;
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
declare function SplitViewHost(props: SplitViewHostProps): React.JSX.Element;
export default SplitViewHost;
//# sourceMappingURL=SplitViewHost.d.ts.map