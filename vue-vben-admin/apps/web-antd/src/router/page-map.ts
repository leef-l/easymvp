// @ts-nocheck

import type { ComponentRecordType } from '@vben/types';

const guardFullBuild = process.env.EASYMVP_WEB_ANTD_GUARD_FULL_BUILD === '1';
const pageMap = import.meta.glob(
  guardFullBuild
    ? ['../views/**/*.vue', '!../views/mvp/workflow/**/*.vue']
    : '../views/**/*.vue',
) as ComponentRecordType;

export { pageMap };
