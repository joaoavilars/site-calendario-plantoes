export interface Membro {
  id: number
  nome: string
  cor: string
  mencao_telegram: string
  ativo: boolean
  criado_em: string
}

export interface Regra {
  id: number
  membro_id: number
  dia_semana: number
  intervalo_semanas: number
  inicio: string
  fim: string
  vigente_de: string
  vigente_ate: string | null
}

export type TipoExcecao =
  | 'substituicao'
  | 'troca'
  | 'extra'
  | 'ausencia'
  | 'feriado'
  | 'edicao_dia'
  | 'ferias'
export type StatusExcecao = 'pendente' | 'confirmado' | 'rejeitado' | 'cancelado'

export interface Excecao {
  id: number
  data: string
  tipo: TipoExcecao
  membro_original_id: number | null
  membro_substituto_id: number | null
  data_troca: string | null
  data_fim: string | null
  inicio: string
  fim: string
  slots_json: string | null
  status: StatusExcecao
  observacao: string
  criado_em: string
  confirmado_em: string | null
}

export interface Slot {
  membro_id: number
  nome: string
  cor: string
  inicio: string
  fim: string
  origem: 'rodizio' | 'substituicao' | 'troca' | 'extra' | 'edicao' | 'ferias'
}

export interface Dia {
  data: string
  slots: Slot[]
  pendentes: Excecao[]
  descoberto: boolean
  sem_plantao: boolean
  motivo: string
  ferias: Array<{ membro_id: number; nome: string }>
}

export interface Horas {
  membro_id: number
  nome: string
  cor: string
  total_minutos: number
  total_formatado: string
}

export interface Calendario {
  ano: number
  mes: number
  dias: Dia[]
  horas: Horas[]
}

export interface PreviewDia {
  data: string
  antes: Slot[]
  depois: Slot[]
}

export interface Preview {
  dias: PreviewDia[]
  horas_antes: Horas[]
  horas_depois: Horas[]
}

export interface AuditEntry {
  id: number
  entidade: string
  entidade_id: number
  acao: string
  dados: unknown
  criado_em: string
}
