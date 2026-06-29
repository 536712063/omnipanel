<template>
  <div class="main-layout" :class="{ 'sidebar-collapsed': settings.sidebarCollapsed }">
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <div class="logo">
          <span class="logo-icon">⚡</span>
          <span v-show="!settings.sidebarCollapsed" class="logo-text">OmniPanel</span>
        </div>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="settings.sidebarCollapsed"
        :background-color="'var(--bg-sidebar)'"
        :text-color="'#c0c4cc'"
        :active-text-color="'#409eff'"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <el-menu-item index="/docker">
          <el-icon><Box /></el-icon>
          <template #title>Docker 管理</template>
        </el-menu-item>

        <el-menu-item index="/ssh">
          <el-icon><Connection /></el-icon>
          <template #title>SSH 工具</template>
        </el-menu-item>

        <el-menu-item index="/frp">
          <el-icon><Link /></el-icon>
          <template #title>FRP 穿透</template>
        </el-menu-item>

        <el-menu-item index="/servers">
          <el-icon><Monitor /></el-icon>
          <template #title>多机管理</template>
        </el-menu-item>

        <el-menu-item index="/cloud">
          <el-icon><Cloudy /></el-icon>
          <template #title>云存储</template>
        </el-menu-item>

        <el-menu-item index="/git">
          <el-icon><FolderOpened /></el-icon>
          <template #title>代码仓库</template>
        </el-menu-item>

        <el-menu-item index="/ai">
          <el-icon><Cpu /></el-icon>
          <template #title>AI 助手</template>
        </el-menu-item>

        <el-menu-item index="/translate">
          <el-icon><Document /></el-icon>
          <template #title>汉化工具</template>
        </el-menu-item>

        <el-menu-item index="/game">
          <el-icon><VideoGame /></el-icon>
          <template #title>七日杀管理</template>
        </el-menu-item>

        <el-menu-item index="/browser">
          <el-icon><ChromeFilled /></el-icon>
          <template #title>内置浏览器</template>
        </el-menu-item>

        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>系统设置</template>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <div class="version-info" v-show="!settings.sidebarCollapsed">
          v1.0.0
        </div>
      </div>
    </aside>

    <main class="main-content">
      <header class="content-header">
        <div class="header-left">
          <el-button
            :icon="settings.sidebarCollapsed ? 'Expand' : 'Fold'"
            @click="settings.toggleSidebar()"
            text
          />
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">OmniPanel</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-tooltip :content="settings.isDark ? '切换到亮色模式' : '切换到暗色模式'">
            <el-button
              :icon="settings.isDark ? 'Sunny' : 'Moon'"
              @click="toggleTheme"
              text
              circle
            />
          </el-tooltip>
          <el-tooltip content="通知中心">
            <el-badge :value="3" :hidden="false">
              <el-button icon="Bell" text circle />
            </el-badge>
          </el-tooltip>
        </div>
      </header>

      <div class="content-body">
        <router-view v-slot="{ Component, route }">
          <transition name="fade" mode="out-in">
            <keep-alive :include="cachedViews">
              <component :is="Component" :key="route.path" />
            </keep-alive>
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'

const route = useRoute()
const settings = useSettingsStore()

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => route.meta?.title as string || '')
const sidebarWidth = computed(() =>
  settings.sidebarCollapsed ? 64 : settings.sidebarWidth
)

const cachedViews = ['Dashboard', 'Docker', 'SSH', 'FRP', 'Servers', 'Cloud', 'Git', 'AI', 'Translate', 'Game', 'Settings', 'Browser']

function toggleTheme() {
  const next = settings.isDark ? 'light' : 'dark'
  settings.setTheme(next)
}
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
  flex-shrink: 0;
  border-right: 1px solid rgba(255,255,255,0.05);
}

.sidebar-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.logo-icon {
  font-size: 22px;
  flex-shrink: 0;
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: #e5eaf3;
  white-space: nowrap;
  background: linear-gradient(135deg, #409eff, #67c23a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  border-right: none !important;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid rgba(255,255,255,0.08);
}

.version-info {
  font-size: 11px;
  color: #606266;
  text-align: center;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.content-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.content-body {
  flex: 1;
  overflow: auto;
  padding: 20px;
  background: var(--bg-secondary);
}
</style>
