/// <reference types="react-native/types/modules/Codegen" />
import type { ViewProps } from 'react-native';
import { DirectEventHandler, Int32 } from 'react-native/Libraries/Types/CodegenTypes';
export type GenericEmptyEvent = Readonly<{}>;
export interface NativeProps extends ViewProps {
    maxLifecycleState: Int32;
    screenKey: string;
    onWillAppear?: DirectEventHandler<GenericEmptyEvent>;
    onDidAppear?: DirectEventHandler<GenericEmptyEvent>;
    onWillDisappear?: DirectEventHandler<GenericEmptyEvent>;
    onDidDisappear?: DirectEventHandler<GenericEmptyEvent>;
}
declare const _default: import("react-native/Libraries/Utilities/codegenNativeComponent").NativeComponentType<NativeProps>;
export default _default;
//# sourceMappingURL=StackScreenNativeComponent.d.ts.map