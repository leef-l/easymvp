<script setup lang="ts">
import { ref } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import { Empty } from 'ant-design-vue';

const route = useRoute();
const workflowRunId = ref((route.query.workflowRunId as string) ?? '');
</script>

<template>
  <Page auto-content-height>
    <Empty
      description="Workflow Timeline — 工作流时间线（M5 阶段实现）"
      class="mt-20"
    />
  </Page>
</template>
