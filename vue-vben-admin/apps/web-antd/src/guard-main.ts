import { createApp, h } from 'vue';

const GuardApp = {
  name: 'EasyMvpGuardApp',
  render() {
    return h('main', { class: 'easy-mvp-guard-app' }, 'EasyMVP guard build');
  },
};

createApp(GuardApp).mount('#app');
