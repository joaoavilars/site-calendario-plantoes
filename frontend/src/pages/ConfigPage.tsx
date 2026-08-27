import { useState } from 'react'
import { api } from '../api/client'

const inputCls =
  'w-full rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700'

export default function ConfigPage({ usuario }: { usuario: string }) {
  const [atual, setAtual] = useState('')
  const [nova, setNova] = useState('')
  const [confirmar, setConfirmar] = useState('')
  const [erro, setErro] = useState('')
  const [msg, setMsg] = useState('')
  const [salvando, setSalvando] = useState(false)

  const alterar = async () => {
    setErro('')
    setMsg('')
    if (nova !== confirmar) {
      setErro('A confirmação não confere com a nova senha.')
      return
    }
    if (nova.length < 6) {
      setErro('A nova senha deve ter pelo menos 6 caracteres.')
      return
    }
    setSalvando(true)
    try {
      await api.alterarSenha(atual, nova)
      setMsg('Senha alterada com sucesso!')
      setAtual('')
      setNova('')
      setConfirmar('')
    } catch (e) {
      setErro((e as Error).message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <div className="mx-auto max-w-md space-y-4">
      {erro && <div className="rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>}
      {msg && <div className="rounded bg-green-100 px-4 py-2 text-green-800 dark:bg-green-950 dark:text-green-200">{msg}</div>}

      <div className="rounded-lg bg-white p-6 shadow dark:bg-slate-800">
        <h2 className="mb-1 font-bold">⚙️ Configurações do administrador</h2>
        <p className="mb-4 text-sm text-slate-500 dark:text-slate-400">
          Usuário: <span className="font-medium">{usuario}</span>
        </p>

        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault()
            alterar()
          }}
        >
          <div>
            <label className="mb-1 block text-sm font-medium">Senha atual</label>
            <input
              type="password"
              value={atual}
              onChange={(e) => setAtual(e.target.value)}
              autoComplete="current-password"
              className={inputCls}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Nova senha</label>
            <input
              type="password"
              value={nova}
              onChange={(e) => setNova(e.target.value)}
              autoComplete="new-password"
              className={inputCls}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Confirmar nova senha</label>
            <input
              type="password"
              value={confirmar}
              onChange={(e) => setConfirmar(e.target.value)}
              autoComplete="new-password"
              className={inputCls}
            />
          </div>
          <button
            type="submit"
            disabled={salvando || !atual || !nova || !confirmar}
            className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            Alterar senha
          </button>
        </form>
        <p className="mt-3 text-xs text-slate-500 dark:text-slate-400">
          Ao alterar a senha, as outras sessões abertas são desconectadas.
        </p>
      </div>
    </div>
  )
}
