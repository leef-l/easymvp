// @ts-nocheck

import type { ComponentRecordType } from '@vben/types';

const pageMap = import.meta.glob('../views/**/*.vue') as ComponentRecordType;

export { pageMap };
