/**
 * Vendored from gridstack.js official Vue 3 wrapper (MIT License).
 * Source: https://github.com/gridstack/gridstack.js/tree/master/vue/projects/lib/src
 *
 * npm `gridstack@13.0.0` does not ship `dist/vue` yet, so the wrapper sources
 * are kept here verbatim (minus `base-widget.ts`, Options API only) to ease
 * future upstream syncs.
 */
export { GridStack, GridStackComponent, type ComponentMap } from "./gridstack";

export { GridStackItem } from "./gridstack-item";

export type {
  GridStackHostApi,
  GridStackWidget,
  GridStackWidgetProps,
  GridStackNode,
  GridStackOptions,
  GridHTMLElement,
  GridItemHTMLElement,
} from "./types";

export {
  installGridStackVueCallbacks,
  gsCreateVueComponents,
  gsSaveAdditionalVueInfo,
  gsUpdateVueComponents,
} from "./registry";

export {
  useGridStack,
  useGridStackItem,
  useWidgetSerializer,
  type UseWidgetSerializerOptions,
  type UseGridStackItemResult,
} from "./composables";

export {
  GS_CONTEXT_KEY,
  GS_ITEM_CONTEXT_KEY,
  useGsContext,
  useGsItemContext,
  type GsContext,
  type GsItemContext,
} from "./gridstack-context";
