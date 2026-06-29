import express from 'express'
import cors from 'cors'
import { createServer } from 'http'
import { WebSocketServer, WebSocket } from 'ws'
import Database from 'better-sqlite3'
import { join } from 'path'
import { exec } from 'child_process'
import * as os from 'os'

const app = express()
const PORT = parseInt(process.env.PORT || '3001')

app.use(cors())
app.use(express.json())

const db = new Database(join(process.cwd(), 'data', 'omnipanel.db'))
db.pragma('journal_mode = WAL')

db.exec(`
  CREATE TABLE IF NOT EXISTS hosts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER DEFAULT 22,
    username TEXT DEFAULT 'root',
    auth_type TEXT DEFAULT 'password',
    password TEXT,
    private_key TEXT,
    "group" TEXT DEFAULT '默认',
    tags TEXT DEFAULT '[]',
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS frp_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    mode TEXT DEFAULT 'frpc',
    content TEXT DEFAULT '',
    enabled INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    ip TEXT NOT NULL,
    port INTEGER DEFAULT 22,
    username TEXT DEFAULT 'root',
    password TEXT,
    "group" TEXT DEFAULT '默认',
    tags TEXT DEFAULT '[]',
    status TEXT DEFAULT 'offline',
    created_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS chat_history (
    id TEXT PRIMARY KEY,
    title TEXT,
    model TEXT,
    messages TEXT DEFAULT '[]',
    created_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    provider TEXT UNIQUE NOT NULL,
    key_value TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS shortcuts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    command TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS backup_tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    target TEXT,
    schedule TEXT,
    last_run TEXT,
    status TEXT DEFAULT 'pending',
    created_at TEXT DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  );

  INSERT OR IGNORE INTO settings (key, value) VALUES ('theme', 'dark');
  INSERT OR IGNORE INTO settings (key, value) VALUES ('language', 'zh-CN');
`)

const server = createServer(app)
const wss = new WebSocketServer({ server, path: '/ws' })

wss.on('connection', (ws: WebSocket) => {
  ws.send(JSON.stringify({ type: 'connected', message: 'WebSocket connected' }))

  ws.on('message', (data) => {
    try {
      const msg = JSON.parse(data.toString())
      if (msg.type === 'terminal') {
        ws.send(JSON.stringify({ type: 'terminal-output', data: `Command output: ${msg.command}` }))
      }
    } catch {
      ws.send(JSON.stringify({ type: 'error', message: 'Invalid message format' }))
    }
  })
})

function broadcast(data: any) {
  wss.clients.forEach(client => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(data))
    }
  })
}

app.get('/api/status', (_req, res) => {
  res.json({
    status: 'ok',
    version: '1.0.0',
    uptime: process.uptime(),
    memory: process.memoryUsage(),
    platform: os.platform()
  })
})

app.get('/api/system/info', (_req, res) => {
  const cpus = os.cpus()
  const totalMem = os.totalmem()
  const freeMem = os.freemem()
  res.json({
    platform: os.platform(),
    arch: os.arch(),
    hostname: os.hostname(),
    release: os.release(),
    uptime: os.uptime(),
    cpuModel: cpus[0]?.model || 'Unknown',
    cpuCores: cpus.length,
    totalMemory: totalMem,
    freeMemory: freeMem,
    usedMemory: totalMem - freeMem
  })
})

app.get('/api/hosts', (_req, res) => {
  const hosts = db.prepare('SELECT * FROM hosts ORDER BY created_at DESC').all()
  const parsed = hosts.map((h: any) => ({
    ...h,
    tags: JSON.parse(h.tags || '[]')
  }))
  res.json(parsed)
})

app.post('/api/hosts', (req, res) => {
  const { id, name, host, port, username, authType, password, privateKey, group, tags } = req.body
  const tagStr = JSON.stringify(tags || [])
  db.prepare(
    'INSERT OR REPLACE INTO hosts (id, name, host, port, username, auth_type, password, private_key, "group", tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)'
  ).run(id, name, host, port || 22, username || 'root', authType || 'password', password, privateKey, group || '默认', tagStr)
  res.json({ success: true })
})

app.delete('/api/hosts/:id', (req, res) => {
  db.prepare('DELETE FROM hosts WHERE id = ?').run(req.params.id)
  res.json({ success: true })
})

app.get('/api/frp/configs', (_req, res) => {
  const configs = db.prepare('SELECT * FROM frp_configs ORDER BY created_at DESC').all()
  res.json(configs)
})

app.post('/api/frp/configs', (req, res) => {
  const { id, name, mode, content, enabled } = req.body
  db.prepare(
    'INSERT OR REPLACE INTO frp_configs (id, name, mode, content, enabled) VALUES (?, ?, ?, ?, ?)'
  ).run(id, name, mode || 'frpc', content || '', enabled ? 1 : 0)
  res.json({ success: true })
})

app.delete('/api/frp/configs/:id', (req, res) => {
  db.prepare('DELETE FROM frp_configs WHERE id = ?').run(req.params.id)
  res.json({ success: true })
})

app.get('/api/servers', (_req, res) => {
  const servers = db.prepare('SELECT * FROM servers ORDER BY created_at DESC').all()
  res.json(servers)
})

app.post('/api/servers', (req, res) => {
  const { id, name, ip, port, username, password, group, tags } = req.body
  db.prepare(
    'INSERT OR REPLACE INTO servers (id, name, ip, port, username, password, "group", tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?)'
  ).run(id, name, ip, port || 22, username || 'root', password, group || '默认', JSON.stringify(tags || []))
  res.json({ success: true })
})

app.delete('/api/servers/:id', (req, res) => {
  db.prepare('DELETE FROM servers WHERE id = ?').run(req.params.id)
  res.json({ success: true })
})

app.get('/api/ai/history', (_req, res) => {
  const history = db.prepare('SELECT id, title, model, created_at FROM chat_history ORDER BY created_at DESC').all()
  res.json(history)
})

app.post('/api/ai/history', (req, res) => {
  const { id, title, model, messages } = req.body
  db.prepare(
    'INSERT OR REPLACE INTO chat_history (id, title, model, messages) VALUES (?, ?, ?, ?)'
  ).run(id, title, model, JSON.stringify(messages || []))
  res.json({ success: true })
})

app.get('/api/ai/keys', (_req, res) => {
  const keys = db.prepare('SELECT id, provider, created_at FROM api_keys ORDER BY created_at DESC').all()
  res.json(keys)
})

app.post('/api/ai/keys', (req, res) => {
  const { id, provider, keyValue } = req.body
  db.prepare('INSERT OR REPLACE INTO api_keys (id, provider, key_value) VALUES (?, ?, ?)').run(id, provider, keyValue)
  res.json({ success: true })
})

app.get('/api/shortcuts', (_req, res) => {
  const shortcuts = db.prepare('SELECT * FROM shortcuts ORDER BY created_at DESC').all()
  res.json(shortcuts)
})

app.post('/api/shortcuts', (req, res) => {
  const { id, name, command } = req.body
  db.prepare('INSERT OR REPLACE INTO shortcuts (id, name, command) VALUES (?, ?, ?)').run(id, name, command)
  res.json({ success: true })
})

app.get('/api/settings', (_req, res) => {
  const rows = db.prepare('SELECT * FROM settings').all() as { key: string; value: string }[]
  const settings: Record<string, string> = {}
  rows.forEach(r => { settings[r.key] = r.value })
  res.json(settings)
})

app.post('/api/settings', (req, res) => {
  const { key, value } = req.body
  db.prepare('INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)').run(key, value)
  res.json({ success: true })
})

app.post('/api/docker/containers', async (_req, res) => {
  exec('docker ps -a --format "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}"', (error, stdout) => {
    if (error) {
      return res.json({ error: 'Docker not available', containers: [] })
    }
    const containers = stdout.trim().split('\n').filter(Boolean).map(line => {
      const [id, name, image, status, ports] = line.split('|')
      return { id, name, image, status, ports: ports || '-' }
    })
    res.json({ containers })
  })
})

app.get('/api/docker/images', async (_req, res) => {
  exec('docker images --format "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}"', (error, stdout) => {
    if (error) {
      return res.json({ error: 'Docker not available', images: [] })
    }
    const images = stdout.trim().split('\n').filter(Boolean).map(line => {
      const [id, repository, tag, size, created] = line.split('|')
      return { id, repository, tag, size, created }
    })
    res.json({ images })
  })
})

app.post('/api/translate', async (req, res) => {
  const { text, source, target, engine } = req.body
  res.json({
    translation: `[${engine || 'auto'}] ${text}`,
    source: source || 'auto',
    target: target || 'zh-CN'
  })
})

app.get('/api/backup/tasks', (_req, res) => {
  const tasks = db.prepare('SELECT * FROM backup_tasks ORDER BY created_at DESC').all()
  res.json(tasks)
})

app.post('/api/backup/tasks', (req, res) => {
  const { id, name, source, target, schedule } = req.body
  db.prepare(
    'INSERT OR REPLACE INTO backup_tasks (id, name, source, target, schedule) VALUES (?, ?, ?, ?, ?)'
  ).run(id, name, source, target, schedule)
  res.json({ success: true })
})

setInterval(() => {
  const totalMem = os.totalmem()
  const freeMem = os.freemem()
  const usedMem = totalMem - freeMem
  broadcast({
    type: 'system-stats',
    data: {
      cpu: (os.loadavg()[0] * 100 / os.cpus().length).toFixed(1),
      memory: ((usedMem / totalMem) * 100).toFixed(1),
      timestamp: Date.now()
    }
  })
}, 3000)

server.listen(PORT, () => {
  console.log(`OmniPanel API server running on http://localhost:${PORT}`)
  console.log(`WebSocket server ready on ws://localhost:${PORT}/ws`)
})
