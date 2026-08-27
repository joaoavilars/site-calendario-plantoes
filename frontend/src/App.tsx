import { useEffect, useState } from 'react'
import { api, getToken, setToken } from './api/client'
import CalendarioPage from './pages/CalendarioPage'
import IntegrantesPage from './pages/IntegrantesPage'
import RodizioPage from './pages/RodizioPage'
import HistoricoPage from './pages/HistoricoPage'
import ConfigPage from './pages/ConfigPage'

const abasAdmin = [
  { id: 'integrantes', rotulo: '👥 Integrantes' },
  { id: 'rodizio', rotulo: '🔁 Rodízio' },
  { id: 'historico', rotulo: '📜 Histórico' },
] as const

type Aba = 'calendario' | 'integrantes' | 'rodizio' | 'historico' | 'config'

function temaInicial(): 'dark' | 'light' {
  try {
    return localStorage.getItem('tema') === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

export default function App() {
  const [aba, setAba] = useState<Aba>('calendario')
  const [tema, setTema] = useState<'dark' | 'light'>(temaInicial)
  const [logado, setLogado] = useState<string | null>(null)
  const [usuario, setUsuario] = useState('')
  const [senha, setSenha] = useState('')
  const [erroLogin, setErroLogin] = useState('')
  const [entrando, setEntrando] = useState(false)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', tema === 'dark')
    try {
      localStorage.setItem('tema', tema)
    } catch {
      // sem localStorage — só não persiste
    }
  }, [tema])

  // valida sessão salva ao abrir
  useEffect(() => {
    if (!getToken()) return
    api
      .me()
      .then((r) => setLogado(r.usuario))
      .catch(() => setToken(null))
  }, [])

  const entrar = async () => {
    setErroLogin('')
    setEntrando(true)
    try {
      const r = await api.login(usuario, senha)
      setToken(r.token)
      setLogado(r.usuario)
      setUsuario('')
      setSenha('')
    } catch (e) {
      setErroLogin((e as Error).message)
    } finally {
      setEntrando(false)
    }
  }

  const sair = async () => {
    try {
      await api.logout()
    } catch {
      // sessão pode já ter expirado — segue o logout local
    }
    setToken(null)
    setLogado(null)
    setAba('calendario')
  }

  return (
    <div className="min-h-screen">
      <header className="bg-slate-800 text-white shadow dark:bg-slate-950">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3">
          <h1 className="text-lg font-bold">Calendário de Plantões</h1>
          <nav className="flex gap-1">
            <button
              onClick={() => setAba('calendario')}
              className={`rounded px-3 py-1.5 text-sm transition ${
                aba === 'calendario' ? 'bg-white/20 font-semibold' : 'hover:bg-white/10'
              }`}
            >
              📅 Calendário
            </button>
            {logado &&
              abasAdmin.map((a) => (
                <button
                  key={a.id}
                  onClick={() => setAba(a.id)}
                  className={`rounded px-3 py-1.5 text-sm transition ${
                    aba === a.id ? 'bg-white/20 font-semibold' : 'hover:bg-white/10'
                  }`}
                >
                  {a.rotulo}
                </button>
              ))}
          </nav>

          <div className="ml-auto flex flex-wrap items-center gap-2">
            {!logado ? (
              <form
                className="flex items-center gap-2"
                onSubmit={(e) => {
                  e.preventDefault()
                  entrar()
                }}
              >
                <input
                  value={usuario}
                  onChange={(e) => setUsuario(e.target.value)}
                  placeholder="usuário"
                  autoComplete="username"
                  className="w-28 rounded border border-white/20 bg-white/10 px-2 py-1 text-sm placeholder-white/50 focus:border-white/50 focus:outline-none"
                />
                <input
                  type="password"
                  value={senha}
                  onChange={(e) => setSenha(e.target.value)}
                  placeholder="senha"
                  autoComplete="current-password"
                  className="w-28 rounded border border-white/20 bg-white/10 px-2 py-1 text-sm placeholder-white/50 focus:border-white/50 focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={entrando || !usuario || !senha}
                  className="rounded bg-blue-600 px-3 py-1 text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  Entrar
                </button>
              </form>
            ) : (
              <>
                <span className="text-sm text-white/70">👤 {logado}</span>
                <button
                  onClick={sair}
                  className="rounded px-2 py-1.5 text-sm hover:bg-white/10"
                  title="Sair"
                >
                  Sair
                </button>
              </>
            )}
            <button
              onClick={() => setTema(tema === 'dark' ? 'light' : 'dark')}
              title={tema === 'dark' ? 'Mudar para tema claro' : 'Mudar para tema escuro'}
              className="rounded px-2 py-1.5 text-sm hover:bg-white/10"
            >
              {tema === 'dark' ? '☀️' : '🌙'}
            </button>
            {logado && (
              <button
                onClick={() => setAba('config')}
                title="Configurações do administrador"
                className={`rounded px-2 py-1.5 text-sm transition ${
                  aba === 'config' ? 'bg-white/20' : 'hover:bg-white/10'
                }`}
              >
                ⚙️
              </button>
            )}
          </div>
        </div>
        {erroLogin && !logado && (
          <div className="bg-red-600/90 px-4 py-1.5 text-center text-sm">{erroLogin}</div>
        )}
      </header>
      <main className="mx-auto max-w-7xl px-4 py-6">
        {aba === 'calendario' && <CalendarioPage logado={!!logado} />}
        {logado && aba === 'integrantes' && <IntegrantesPage />}
        {logado && aba === 'rodizio' && <RodizioPage />}
        {logado && aba === 'historico' && <HistoricoPage />}
        {logado && aba === 'config' && <ConfigPage usuario={logado} />}
      </main>
    </div>
  )
}
