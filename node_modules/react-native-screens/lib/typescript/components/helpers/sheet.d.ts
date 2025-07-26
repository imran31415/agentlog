import { ScreenProps } from '../../types';
export declare const SHEET_FIT_TO_CONTENTS: number[];
export declare const SHEET_COMPAT_LARGE: number[];
export declare const SHEET_COMPAT_MEDIUM: number[];
export declare const SHEET_COMPAT_ALL: number[];
export declare const SHEET_DIMMED_ALWAYS = -1;
export declare function assertDetentsArrayIsSorted(array: number[]): void;
export declare function resolveSheetAllowedDetents(allowedDetentsCompat: ScreenProps['sheetAllowedDetents']): number[];
export declare function resolveSheetLargestUndimmedDetent(lud: ScreenProps['sheetLargestUndimmedDetentIndex'], lastDetentIndex: number): number;
export declare function resolveSheetInitialDetentIndex(index: ScreenProps['sheetInitialDetentIndex'], lastDetentIndex: number): number;
//# sourceMappingURL=sheet.d.ts.map