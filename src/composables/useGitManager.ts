import { ref } from 'vue'
import type { GitRepo, GitCloneRequest, GitBranch, GitCommit, GitStatusItem } from '@/wails/runtime'
import * as api from '@/wails/runtime'

export function useGitManager() {
  const repos = ref<GitRepo[]>([])
  const selectedRepoId = ref('')
  const branches = ref<GitBranch[]>([])
  const commits = ref<GitCommit[]>([])
  const statusItems = ref<GitStatusItem[]>([])
  const stats = ref<Record<string, unknown>>({})
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadRepos() { repos.value = await api.gitListRepos() }

  async function addLocalRepo(path: string) { await api.gitAddLocalRepo(path); await loadRepos() }
  async function cloneRepo(req: GitCloneRequest) { await api.gitCloneRepo(req); await loadRepos() }

  async function selectRepo(id: string) {
    selectedRepoId.value = id
    loading.value = true
    try {
      const [b, c, s, st] = await Promise.all([
        api.gitBranches(id),
        api.gitLog(id, 50),
        api.gitStatus(id),
        api.gitGetRepoStats(id),
      ])
      branches.value = b; commits.value = c; statusItems.value = s; stats.value = st
    } catch (e: any) { error.value = e?.message } finally { loading.value = false }
  }

  async function pull() { await api.gitPull(selectedRepoId.value); await selectRepo(selectedRepoId.value) }
  async function push() { await api.gitPush(selectedRepoId.value); await selectRepo(selectedRepoId.value) }
  async function commit(message: string) { await api.gitCommit(selectedRepoId.value, message); await selectRepo(selectedRepoId.value) }
  async function checkout(branch: string) { await api.gitCheckout(selectedRepoId.value, branch); await selectRepo(selectedRepoId.value) }
  async function createBranch(name: string) { await api.gitCreateBranch(selectedRepoId.value, name); await selectRepo(selectedRepoId.value) }

  return { repos, selectedRepoId, branches, commits, statusItems, stats, loading, error,
    loadRepos, addLocalRepo, cloneRepo, selectRepo, pull, push, commit, checkout, createBranch }
}
