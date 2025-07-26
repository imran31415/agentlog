import { controlEdgeToEdgeValues, isEdgeToEdge } from 'react-native-is-edge-to-edge';
export const EDGE_TO_EDGE = isEdgeToEdge();
export function transformEdgeToEdgeProps(props) {
  const {
    // Filter out edge-to-edge related props
    statusBarColor,
    statusBarTranslucent,
    navigationBarColor,
    navigationBarTranslucent,
    ...rest
  } = props;
  if (__DEV__) {
    controlEdgeToEdgeValues({
      statusBarColor,
      statusBarTranslucent,
      navigationBarColor,
      navigationBarTranslucent
    });
  }
  return rest;
}
//# sourceMappingURL=edge-to-edge.js.map