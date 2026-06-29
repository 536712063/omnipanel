import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' }
      },
      {
        path: 'docker',
        name: 'Docker',
        component: () => import('@/views/docker/DockerCenter.vue'),
        meta: { title: 'Docker 管理', icon: 'Box' }
      },
      {
        path: 'ssh',
        name: 'SSH',
        component: () => import('@/views/ssh/SSHTool.vue'),
        meta: { title: 'SSH 工具', icon: 'Connection' }
      },
      {
        path: 'frp',
        name: 'FRP',
        component: () => import('@/views/frp/FRPManager.vue'),
        meta: { title: 'FRP 穿透', icon: 'Link' }
      },
      {
        path: 'servers',
        name: 'Servers',
        component: () => import('@/views/servers/MultiServer.vue'),
        meta: { title: '多机管理', icon: 'Monitor' }
      },
      {
        path: 'cloud',
        name: 'Cloud',
        component: () => import('@/views/cloud/CloudStorage.vue'),
        meta: { title: '云存储', icon: 'Cloudy' }
      },
      {
        path: 'git',
        name: 'Git',
        component: () => import('@/views/git/CodeRepo.vue'),
        meta: { title: '代码仓库', icon: 'FolderOpened' }
      },
      {
        path: 'ai',
        name: 'AI',
        component: () => import('@/views/ai/AIAssistant.vue'),
        meta: { title: 'AI 助手', icon: 'Cpu' }
      },
      {
        path: 'translate',
        name: 'Translate',
        component: () => import('@/views/translate/TranslateTool.vue'),
        meta: { title: '汉化工具', icon: 'Document' }
      },
      {
        path: 'game',
        name: 'Game',
        component: () => import('@/views/game/GameManager.vue'),
        meta: { title: '七日杀', icon: 'VideoGame' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/Settings.vue'),
        meta: { title: '系统设置', icon: 'Setting' }
      },
      {
        path: 'browser',
        name: 'Browser',
        component: () => import('@/views/browser/BuiltInBrowser.vue'),
        meta: { title: '内置浏览器', icon: 'ChromeFilled' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
