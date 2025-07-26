import React from 'react';
import type { ViewProps } from 'react-native';
import type { NativeProps } from '../../fabric/gamma/StackScreenNativeComponent';
export declare const StackScreenLifecycleState: {
    readonly INITIAL: 0;
    readonly DETACHED: 1;
    readonly ATTACHED: 2;
};
export type StackScreenNativeProps = NativeProps & {
    maxLifecycleState: (typeof StackScreenLifecycleState)[keyof typeof StackScreenLifecycleState];
};
type StackScreenProps = {
    children?: ViewProps['children'];
    onPop?: (screenKey: string) => void;
} & StackScreenNativeProps;
/**
 * EXPERIMENTAL API, MIGHT CHANGE W/O ANY NOTICE
 */
declare function StackScreen({ children, maxLifecycleState, screenKey, onWillAppear, onWillDisappear, onDidAppear, onDidDisappear, onPop, }: StackScreenProps): React.JSX.Element;
export default StackScreen;
//# sourceMappingURL=StackScreen.d.ts.map