import * as workflowApi from '../api/mvp/workflow';

const workflowViews = import.meta.glob('../views/mvp/workflow/**/*.vue', {
  eager: true,
});
const workflowScripts = import.meta.glob('../views/mvp/workflow/**/*.{ts,tsx}', {
  eager: true,
});

const workflowBundleVerify = {
  workflowApi,
  workflowScripts,
  workflowViews,
};

export { workflowBundleVerify };
export default workflowBundleVerify;
