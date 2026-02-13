import apiClient from '@/api/client'

export interface ResellerAnnouncement {
  id: number
  title: string
  content: string
  status: string // draft | active | archived
  starts_at?: string
  ends_at?: string
  created_at: string
  updated_at: string
  owner_id?: number
}

export interface ResellerAnnouncementListResponse {
  items: ResellerAnnouncement[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface CreateAnnouncementRequest {
  title: string
  content: string
  status: string
  starts_at?: string
  ends_at?: string
}

export interface UpdateAnnouncementRequest {
  title?: string
  content?: string
  status?: string
  starts_at?: string | null
  ends_at?: string | null
}

export const resellerAnnouncementAPI = {
  async list(params: { page?: number; page_size?: number; status?: string }) {
    const { data } = await apiClient.get<ResellerAnnouncementListResponse>('/reseller/announcements', { params })
    return data
  },
  async create(requestData: CreateAnnouncementRequest) {
    const { data } = await apiClient.post<ResellerAnnouncement>('/reseller/announcements', requestData)
    return data
  },
  async update(id: number, requestData: UpdateAnnouncementRequest) {
    const { data } = await apiClient.put<ResellerAnnouncement>(`/reseller/announcements/${id}`, requestData)
    return data
  },
  async delete(id: number) {
    const { data } = await apiClient.delete(`/reseller/announcements/${id}`)
    return data
  }
}

export default resellerAnnouncementAPI
