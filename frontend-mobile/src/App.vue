<template>
  <router-view />
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  -webkit-tap-highlight-color: transparent;
}

html, body {
  height: 100%;
  overflow: hidden;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Roboto,
               'PingFang SC', 'Microsoft Yahei', sans-serif;
  background-color: var(--color-bg, #F7F8FA);
  color: var(--color-text-primary, #323233);
  font-size: var(--font-size-body, 14px);
  line-height: 1.5;
}

#app {
  height: 100%;
}
</style>
