import apiClient from '@/api/client'

export interface GenerateRedeemCodesRequest {
  count: number
  value: number
}

export interface ResellerRedeemCode {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_by?: number
  used_at?: string
  created_at: string
  owner_id?: number
  user?: { id: number; email: string }
}

export interface ResellerRedeemCodeListResponse {
  items: ResellerRedeemCode[]
  total: number
  page: number
  page_size: number
  pages: number
}

export const resellerRedeemAPI = {
  async generate(requestData: GenerateRedeemCodesRequest) {
    const { data } = await apiClient.post<ResellerRedeemCode[]>('/reseller/redeem/generate', requestData)
    return data
  },
  async list(params: { page?: number; page_size?: number }) {
    const { data } = await apiClient.get<ResellerRedeemCodeListResponse>('/reseller/redeem', { params })
    return data
  },
  async delete(id: number) {
    const { data } = await apiClient.delete(`/reseller/redeem/${id}`)
    return data
  }
}

export default resellerRedeemAPI
