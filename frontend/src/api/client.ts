import type {
  AuditEntry,
  Calendario,
  Excecao,
  Membro,
  Preview,
  Regra,
} from '../types'

export function getToken(): string {
  try {
    return localStorage.getItem('token') ?? ''
  } catch {
    return ''
  }
}

export function setToken(token: string | null) {
  try {
    if (token) localStorage.setItem('token', token)
    else localStorage.removeItem('token')
  } catch {
    // sem localStorage — sessão não persiste entre reloads
  }
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken()
  const res = await fetch(`/api${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...init,
  })
  const body = await res.json().catch(() => null)
  if (!res.ok) {
    throw new Error(body?.erro ?? `Erro HTTP ${res.status}`)
  }
  return body as T
}

export const api = {
  login: (usuario: string, senha: string) =>
    req<{ token: string; usuario: string }>('/login', {
      method: 'POST',
      body: JSON.stringify({ usuario, senha }),
    }),
  logout: () => req<{ ok: boolean }>('/logout', { method: 'POST' }),
  me: () => req<{ usuario: string }>('/me'),
  alterarSenha: (senha_atual: string, senha_nova: string) =>
    req<{ ok: boolean }>('/senha', {
      method: 'POST',
      body: JSON.stringify({ senha_atual, senha_nova }),
    }),

  listarMembros: (todos = false) =>
    req<Membro[]>(`/membros${todos ? '?todos=1' : ''}`),
  criarMembro: (m: Partial<Membro>) =>
    req<Membro>('/membros', { method: 'POST', body: JSON.stringify(m) }),
  atualizarMembro: (id: number, m: Partial<Membro>) =>
    req<Membro>(`/membros/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
  desativarMembro: (id: number) =>
    req<{ ok: boolean }>(`/membros/${id}`, { method: 'DELETE' }),

  listarRegras: (membroId: number) => req<Regra[]>(`/membros/${membroId}/regras`),
  salvarRegras: (
    membroId: number,
    regras: Array<{
      dia_semana: number
      intervalo_semanas?: number
      inicio: string
      fim: string
      vigente_de?: string
    }>,
    vigenteDe?: string,
  ) =>
    req<Regra[]>(`/membros/${membroId}/regras`, {
      method: 'PUT',
      body: JSON.stringify({ regras, vigente_de: vigenteDe }),
    }),

  calendario: (ano: number, mes: number) =>
    req<Calendario>(`/calendario/${ano}/${mes}`),

  criarExcecao: (e: {
    data: string
    tipo: string
    membro_original_id?: number | null
    membro_substituto_id?: number | null
    data_troca?: string | null
    data_fim?: string | null
    inicio?: string
    fim?: string
    slots?: Array<{ membro_id: number; inicio: string; fim: string }>
    observacao?: string
  }) =>
    req<{ excecao: Excecao; preview: Preview }>('/excecoes', {
      method: 'POST',
      body: JSON.stringify(e),
    }),
  confirmarExcecao: (id: number) =>
    req<Excecao>(`/excecoes/${id}/confirmar`, { method: 'POST' }),
  rejeitarExcecao: (id: number) =>
    req<Excecao>(`/excecoes/${id}/rejeitar`, { method: 'POST' }),
  cancelarExcecao: (id: number) =>
    req<Excecao>(`/excecoes/${id}`, { method: 'DELETE' }),

  listarExcecoes: (params: { tipo?: string; membro?: number; status?: string } = {}) => {
    const q = new URLSearchParams(
      Object.fromEntries(
        Object.entries(params)
          .filter(([, v]) => v !== undefined && v !== '')
          .map(([k, v]) => [k, String(v)]),
      ),
    ).toString()
    return req<Excecao[]>(`/excecoes${q ? `?${q}` : ''}`)
  },

  historico: (params: { entidade?: string; de?: string; ate?: string } = {}) => {
    const q = new URLSearchParams(
      Object.fromEntries(Object.entries(params).filter(([, v]) => v)),
    ).toString()
    return req<AuditEntry[]>(`/historico${q ? `?${q}` : ''}`)
  },
  testarAlerta: () => req<{ ok: boolean }>('/alertas/teste', { method: 'POST' }),
}
