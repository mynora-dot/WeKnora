<script setup lang="ts">
import { resolveSiteFiling } from '@/utils/siteFiling'

const filing = resolveSiteFiling(window.__RUNTIME_CONFIG__)
</script>

<template>
  <footer v-if="filing.visible" class="site-filing-footer">
    <a v-if="filing.icp" href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">
      {{ filing.icp }}
    </a>
    <a v-if="filing.police" :href="filing.police.url" target="_blank" rel="noopener noreferrer">
      <img :src="filing.police.icon" alt="" width="20" height="20" />
      <span>{{ filing.police.number }}</span>
    </a>
  </footer>
</template>

<style scoped>
.site-filing-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 8px 20px;
  padding: 14px 20px calc(14px + env(safe-area-inset-bottom, 0px));
  background: var(--td-bg-color-page);
  color: var(--td-text-color-secondary);
  font-size: 13px;
  line-height: 1.6;
  flex-shrink: 0;
}
a { display: inline-flex; align-items: center; gap: 6px; color: inherit; text-decoration: none; max-width: 100%; overflow-wrap: anywhere; }
a:hover { text-decoration: underline; }
img { flex-shrink: 0; }
</style>
